package service_test

import (
	"task217-fermend/internal/model"
	"task217-fermend/internal/service"
	"task217-fermend/internal/store"
	"testing"
)

func TestSealBatchReachesSealedState(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/seal.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchRunning); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchPendingDiagnosis); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchConfirmed); err != nil {
		t.Fatal(err)
	}
	sealed, err := svc.SealBatch(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sealed.Status != model.BatchSealed {
		t.Fatalf("sealed status = %q", sealed.Status)
	}
}
