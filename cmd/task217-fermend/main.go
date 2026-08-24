// Command task217-fermend 是发酵罐代谢终点漂移诊断服务的入口。
//
// 用法：
//
//	./fermend --addr :8080 --db fermend.db        # 启动 HTTP 服务
//	./fermend --smoke-test                         # 端到端自检（Docker 判据）
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"task217-fermend/internal/httpapi"
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/snapshot"
	"task217-fermend/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "fermend.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run end-to-end self test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmoke(); err != nil {
			log.Fatalf("smoke test failed: %v", err)
		}
		fmt.Println("SMOKE TEST PASSED")
		return
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	svc := service.New(s)
	server := httpapi.New(svc)

	log.Printf("fermend listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// runSmoke 完整复现端到端场景：pH 通道滞后导致预测终点提前 → 标记校正 → 重算 →
// 与离线采样一致 → 发布诊断 → 关闭重开验证持久化恢复。
func runSmoke() error {
	dir, err := os.MkdirTemp("", "fermend-smoke-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbPath := filepath.Join(dir, "smoke.db")

	// 第一段：构建场景。
	s, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	svc := service.New(s)

	batch, err := svc.CreateBatch("E.coli-fed-batch-001", "E. coli BL21", "BR-7")
	if err != nil {
		return err
	}
	_, _ = svc.TransitionBatch(batch.ID, model.BatchRunning)

	doCh, err := svc.RegisterChannel(batch.ID, "dissolved_oxygen", model.KindDissolvedOxygen, model.UnitPercent)
	if err != nil {
		return err
	}
	phCh, err := svc.RegisterChannel(batch.ID, "ph", model.KindPH, model.UnitPH)
	if err != nil {
		return err
	}
	feedCh, err := svc.RegisterChannel(batch.ID, "feed", model.KindFeed, model.UnitGramPerL)
	if err != nil {
		return err
	}

	// 阶段标记（拒绝逆序由 stage 模块保证）。
	if _, err := svc.AddStage(batch.ID, "lag", 0, 200); err != nil {
		return err
	}
	if _, err := svc.AddStage(batch.ID, "log", 200, 600); err != nil {
		return err
	}
	if _, err := svc.AddStage(batch.ID, "stationary", 600, 0); err != nil {
		return err
	}

	// 溶氧曲线：先降后升，谷点在 t=500。
	doTimes := []int64{0, 100, 200, 300, 400, 500, 600, 700, 800}
	doVals := []float64{100, 80, 60, 45, 35, 30, 55, 80, 95}
	var doPts []model.SamplePoint
	for i := range doTimes {
		doPts = append(doPts, model.SamplePoint{Seq: int64(i), TUnix: doTimes[i], Value: doVals[i]})
	}
	if err := svc.IngestSamples(batch.ID, doCh.ID, doPts); err != nil {
		return err
	}

	// pH 曲线：t=530 处突变回升（拐点），代表滞后 30 秒的传感器记录。
	phTimes := []int64{0, 100, 200, 300, 400, 500, 530, 600, 700, 800}
	phVals := []float64{7.0, 6.9, 6.8, 6.7, 6.6, 6.5, 8.5, 8.6, 8.7, 8.8}
	var phPts []model.SamplePoint
	for i := range phTimes {
		phPts = append(phPts, model.SamplePoint{Seq: int64(i), TUnix: phTimes[i], Value: phVals[i]})
	}
	if err := svc.IngestSamples(batch.ID, phCh.ID, phPts); err != nil {
		return err
	}

	// 补料曲线：累积补料在 t=500 处速率突变（拐点），之后速率恒定。
	feedTimes := []int64{0, 100, 200, 300, 400, 500, 600, 700, 800}
	feedVals := []float64{0, 1, 2, 3, 4, 9, 14, 19, 24}
	var feedPts []model.SamplePoint
	for i := range feedTimes {
		feedPts = append(feedPts, model.SamplePoint{Seq: int64(i), TUnix: feedTimes[i], Value: feedVals[i]})
	}
	if err := svc.IngestSamples(batch.ID, feedCh.ID, feedPts); err != nil {
		return err
	}

	// 首次诊断：pH 滞后未校正。
	_, v1, err := svc.Diagnose(batch.ID)
	if err != nil {
		return err
	}
	eps1, _ := svc.ListEndpoints(batch.ID, v1)
	phKneeBefore := endpointTBySource(eps1, "knee_ph")

	// 标记 pH 通道滞后段并应用校正（lag=30 秒）。
	seg, err := svc.MarkSegment(batch.ID, phCh.ID, 400, 800, model.SegmentLagging, 30)
	if err != nil {
		return err
	}
	if _, err := svc.ApplyCorrection(seg.ID); err != nil {
		return err
	}

	// 重算：校正后 pH 拐点应前移 30 秒。
	_, v2, err := svc.Diagnose(batch.ID)
	if err != nil {
		return err
	}
	eps2, _ := svc.ListEndpoints(batch.ID, v2)
	phKneeAfter := endpointTBySource(eps2, "knee_ph")
	if phKneeAfter != phKneeBefore-30 {
		return fmt.Errorf("expected pH knee to shift from %d to %d, got %d", phKneeBefore, phKneeBefore-30, phKneeAfter)
	}

	// 离线采样终点 = 500，校正后综合预测应落入容差内。
	_, cmp, err := svc.ResolveWithSample(batch.ID, v2, 500)
	if err != nil {
		return err
	}
	if !cmp.WithinTolerance {
		return fmt.Errorf("expected within tolerance, got deviation %d", cmp.DeviationSecs)
	}

	// 发布诊断快照。
	ev := snapshot.Evidence{
		EndpointT:     cmp.PredictedT,
		Verdict:       cmp.Verdict,
		DeviationSecs: cmp.DeviationSecs,
		Notes:         "pH lag corrected to 30s",
	}
	draft, err := svc.CreateSnapshotDraft(batch.ID, cmp.PredictedT, ev)
	if err != nil {
		return err
	}
	if _, err := svc.PublishSnapshot(draft.ID); err != nil {
		return err
	}

	// 关闭数据库，验证重启恢复。
	if err := s.Close(); err != nil {
		return err
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer s2.Close()
	svc2 := service.New(s2)

	restored, err := svc2.GetBatch(batch.ID)
	if err != nil {
		return err
	}
	if restored.Status != model.BatchConfirmed {
		return fmt.Errorf("batch status not restored: %s", restored.Status)
	}
	published, err := svc2.LatestPublishedSnapshot(batch.ID)
	if err != nil {
		return err
	}
	if published.Version != draft.Version {
		return fmt.Errorf("snapshot version mismatch after restart")
	}

	fmt.Printf("smoke: batch=%d ph_knee before=%d after=%d sampled=500 deviation=%ds verdict=%s snapshot=v%d\n",
		batch.ID, phKneeBefore, phKneeAfter, cmp.DeviationSecs, cmp.Verdict, published.Version)
	return nil
}

// endpointTBySource 从候选列表提取指定来源的终点时刻。
func endpointTBySource(eps []*model.EndpointCandidate, source string) int64 {
	for _, e := range eps {
		if e.Source == source {
			return e.TUnix
		}
	}
	return 0
}
