package service

import (
	"testing"
	"time"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

// seedCombined 在某批次直接写入一条综合预测候选，避免走完整拐点推断流程，
// 便于聚焦验证 ResolveWithSample 的容差边界判定。
func seedCombined(t *testing.T, svc *Service, batchID int64, predictedT int64) int {
	t.Helper()
	v, err := svc.Store.MaxEndpointVersion(batchID)
	if err != nil {
		t.Fatalf("max version: %v", err)
	}
	v++
	if _, err := svc.Store.CreateEndpoint(&model.EndpointCandidate{
		BatchID:   batchID,
		Version:   v,
		Source:    "combined",
		TUnix:     predictedT,
		Status:    model.EndpointPredicted,
		Reason:    "median of channel knees",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create combined: %v", err)
	}
	return v
}

// TestResolveWithSampleToleranceBoundary 验证：离线采样终点与综合预测恰好相差容差
// （默认 60 秒）时判定为 confirmed；超出容差仍判定为 conflict。
// 覆盖 HTTP→service→compare 数值传递链路，回归「边界差值被 +1 放大后误判冲突」。
func TestResolveWithSampleToleranceBoundary(t *testing.T) {
	const predictedT int64 = 500
	const tolerance int64 = 60

	s, err := store.Open(t.TempDir() + "/tol.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := New(s)

	b, err := svc.CreateBatch("boundary-batch", "strain-1", "BR-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchRunning); err != nil {
		t.Fatal(err)
	}
	v := seedCombined(t, svc, b.ID, predictedT)

	if svc.Diagnosis.ToleranceSecs != tolerance {
		t.Fatalf("tolerance = %d, want %d", svc.Diagnosis.ToleranceSecs, tolerance)
	}

	// 恰好相差容差（60 秒）→ confirmed。
	t.Run("exactly_tolerance_is_confirmed", func(t *testing.T) {
		_, cmp, err := svc.ResolveWithSample(b.ID, v, predictedT+tolerance)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if cmp.DeviationSecs != tolerance {
			t.Fatalf("deviation = %d, want %d", cmp.DeviationSecs, tolerance)
		}
		if !cmp.WithinTolerance || cmp.Verdict != "confirmed" {
			t.Fatalf("at tolerance = %#v, want confirmed", cmp)
		}
	})

	// 同样以 -tolerance 方向验证对称性。
	t.Run("exactly_tolerance_negative_is_confirmed", func(t *testing.T) {
		_, cmp, err := svc.ResolveWithSample(b.ID, v, predictedT-tolerance)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if !cmp.WithinTolerance || cmp.Verdict != "confirmed" {
			t.Fatalf("at -tolerance = %#v, want confirmed", cmp)
		}
	})

	// 超出容差（61 秒）→ conflict。
	t.Run("beyond_tolerance_is_conflict", func(t *testing.T) {
		_, cmp, err := svc.ResolveWithSample(b.ID, v, predictedT+tolerance+1)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if cmp.WithinTolerance || cmp.Verdict != "conflict" {
			t.Fatalf("beyond tolerance = %#v, want conflict", cmp)
		}
	})
}
