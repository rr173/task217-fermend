package service_test

import (
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
	"testing"
)

func TestDuplicateSampleKeepsOriginalDeviceSequenceAndValue(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/samples.db")
	if err != nil {
		_, _ = service.New(s).CreateBatch("batch", "strain", "BR")
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
	if err := svc.IngestSamples(b.ID, ch.ID, []model.SamplePoint{{Seq: 42, TUnix: 10, Value: 12}}); err != nil {
		t.Fatal(err)
	}
	if err := svc.IngestSamples(b.ID, ch.ID, []model.SamplePoint{{Seq: 42, TUnix: 10, Value: 99}}); err != nil {
		t.Fatal(err)
	}
	points, err := svc.ListSamples(ch.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].Seq != 42 || points[0].Value != 12 {
		t.Fatalf("stored duplicate result = %#v", points)
	}
}
