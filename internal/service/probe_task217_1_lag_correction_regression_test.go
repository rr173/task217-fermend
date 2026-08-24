package service_test

import (
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
)

func TestAppliedLagMovesPHKneeEarlier(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/lag.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("lag-batch", "strain", "BR-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchRunning); err != nil {
		t.Fatal(err)
	}
	do, err := svc.RegisterChannel(b.ID, "do", model.KindDissolvedOxygen, model.UnitPercent)
	if err != nil {
		t.Fatal(err)
	}
	ph, err := svc.RegisterChannel(b.ID, "ph", model.KindPH, model.UnitPH)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.IngestSamples(b.ID, do.ID, []model.SamplePoint{
		{Seq: 1, TUnix: 400, Value: 50}, {Seq: 2, TUnix: 500, Value: 20}, {Seq: 3, TUnix: 600, Value: 60},
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.IngestSamples(b.ID, ph.ID, []model.SamplePoint{
		{Seq: 1, TUnix: 400, Value: 6.8}, {Seq: 2, TUnix: 500, Value: 6.7}, {Seq: 3, TUnix: 530, Value: 5.0}, {Seq: 4, TUnix: 600, Value: 5.1},
	}); err != nil {
		t.Fatal(err)
	}
	seg, err := svc.EstimateLag(b.ID, do.ID, ph.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ApplyCorrection(seg.ID); err != nil {
		t.Fatal(err)
	}
	endpoints, _, err := svc.Diagnose(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range endpoints {
		if endpoint.Source == "knee_ph" && endpoint.TUnix != 500 {
			t.Fatalf("corrected pH knee = %d, want 500", endpoint.TUnix)
		}
	}
}
