package service_test

import (
	"task217-fermend/internal/service"
	"task217-fermend/internal/snapshot"
	"task217-fermend/internal/store"
	"testing"
)

func TestPublishedSnapshotRetainsEvidenceAfterReopen(t *testing.T) {
	dbPath := t.TempDir() + "/evidence.db"
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	draft, err := svc.CreateSnapshotDraft(b.ID, 500, snapshot.Evidence{EndpointT: 500, Verdict: "confirmed", Notes: "retain me"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishSnapshot(draft.ID); err != nil {
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
	got, err := service.New(s2).LatestPublishedSnapshot(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Evidence == "" {
		t.Fatal("published snapshot evidence was lost after reopen")
	}
}
