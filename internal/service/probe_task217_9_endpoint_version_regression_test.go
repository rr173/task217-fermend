package service_test

import (
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
	"testing"
)

func TestDiagnosisVersionsStartAtOneAndAdvance(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/versions.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	ch, err := svc.RegisterChannel(b.ID, "do", model.KindDissolvedOxygen, model.UnitPercent)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.IngestSamples(b.ID, ch.ID, []model.SamplePoint{{Seq: 1, TUnix: 10, Value: 1}, {Seq: 2, TUnix: 20, Value: 2}}); err != nil {
		t.Fatal(err)
	}
	_, first, err := svc.Diagnose(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := svc.Diagnose(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("diagnosis versions = %d then %d", first, second)
	}
}
