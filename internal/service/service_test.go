package service

import (
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

func TestBatchPersistsAcrossReopen(t *testing.T) {
	dbPath := t.TempDir() + "/service.db"
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	svc := New(s)
	b, err := svc.CreateBatch("batch-1", "strain-1", "BR-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionBatch(b.ID, model.BatchRunning); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	restored, err := New(s2).GetBatch(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Status != model.BatchRunning || restored.Name != "batch-1" {
		t.Fatalf("restored batch = %#v", restored)
	}
}
