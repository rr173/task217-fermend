package diagnosis

import (
	"errors"
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

func TestKneeDetectors(t *testing.T) {
	do := []*model.SamplePoint{{TUnix: 0, Value: 80}, {TUnix: 10, Value: 30}, {TUnix: 20, Value: 60}}
	if got, _, ok := DoReboundTime(do); !ok || got != 10 {
		t.Fatalf("DoReboundTime = (%d, %v), want (10, true)", got, ok)
	}
	ph := []*model.SamplePoint{{TUnix: 0, Value: 7}, {TUnix: 10, Value: 6.9}, {TUnix: 20, Value: 8.5}}
	if got, _, ok := PhKneeTime(ph); !ok || got != 20 {
		t.Fatalf("PhKneeTime = (%d, %v), want (20, true)", got, ok)
	}
	feed := []*model.SamplePoint{{TUnix: 0, Value: 0}, {TUnix: 10, Value: 1}, {TUnix: 20, Value: 5}, {TUnix: 30, Value: 15}}
	if got, _, ok := FeedKneeTime(feed); !ok || got != 30 {
		t.Fatalf("FeedKneeTime = (%d, %v), want (30, true)", got, ok)
	}
}

func TestMedianAndComparison(t *testing.T) {
	if got := MedianTime([]int64{530, 500, 500}); got != 500 {
		t.Fatalf("MedianTime = %d, want 500", got)
	}
	if got := MedianTime([]int64{500, 530}); got != 530 {
		t.Fatalf("even MedianTime = %d, want upper middle 530", got)
	}
	d := &Diagnosis{ToleranceSecs: 30}
	if got := d.Compare(500, 530); !got.WithinTolerance || got.Verdict != "confirmed" || got.DeviationSecs != 30 {
		t.Fatalf("Compare at tolerance = %#v, want confirmed with deviation 30", got)
	}
	if got := d.Compare(500, 531); got.WithinTolerance || got.Verdict != "conflict" {
		t.Fatalf("Compare beyond tolerance = %#v, want conflict", got)
	}
}

// TestInferEndpointsRejectsExcludedChannel 锁定诊断只读未剔除通道：
// 唯一诊断通道被剔除后，InferEndpoints 应拒绝并返回可处理的 ErrNoActiveDiagnosis，
// 而不是仍按剔除通道的曲线生成拐点诊断成功。
func TestInferEndpointsRejectsExcludedChannel(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/excluded.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	b, err := s.CreateBatch(model.NewBatch("b", "s", "BR"))
	if err != nil {
		t.Fatal(err)
	}
	ch, err := (&samplingChannelHelper{s}).register(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	pts := []*model.SamplePoint{
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 0, TUnix: 0, Value: 100},
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 1, TUnix: 100, Value: 30},
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 2, TUnix: 200, Value: 60},
	}
	if err := s.InsertSamplesTx(pts); err != nil {
		t.Fatal(err)
	}

	d := New(s)
	if _, _, err := d.InferEndpoints(b.ID); err != nil {
		t.Fatalf("diagnose before exclude failed: %v", err)
	}
	// 之前的诊断版本不应在剔除后被纳入。
	if err := s.ExcludeChannel(ch.ID); err != nil {
		t.Fatalf("exclude channel: %v", err)
	}
	if ch2, _ := s.GetChannel(ch.ID); ch2.Status != model.ChannelExcluded {
		t.Fatalf("channel status = %q, want excluded (stored)", ch2.Status)
	}
	if _, _, err := d.InferEndpoints(b.ID); err == nil {
		t.Fatal("diagnose after exclude succeeded, want rejection")
	} else if !errors.Is(err, model.ErrNoActiveDiagnosis) {
		t.Fatalf("diagnose after exclude error = %v, want ErrNoActiveDiagnosis", err)
	}
}

type samplingChannelHelper struct {
	store *store.Store
}

func (h *samplingChannelHelper) register(batchID int64) (*model.Channel, error) {
	return h.store.CreateChannel(&model.Channel{
		BatchID: batchID, Name: "do", Kind: model.KindDissolvedOxygen,
		Unit: model.UnitPercent, Status: model.ChannelActive,
	})
}
