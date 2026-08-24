package service_test

import (
	"testing"

	"task217-fermend/internal/service"
	"task217-fermend/internal/snapshot"
	"task217-fermend/internal/store"
)

func TestSupersedePromotesNewSnapshot(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/snapshot.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s)
	b, err := svc.CreateBatch("batch", "strain", "BR")
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.CreateSnapshotDraft(b.ID, 500, snapshot.Evidence{EndpointT: 500, Verdict: "confirmed"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishSnapshot(first.ID); err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateSnapshotDraft(b.ID, 510, snapshot.Evidence{EndpointT: 510, Verdict: "confirmed"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.SupersedeSnapshot(first.ID, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != second.ID || got.Status != "published" {
		t.Fatalf("supersede result = %#v", got)
	}
	latest, err := svc.LatestPublishedSnapshot(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.ID != second.ID {
		t.Fatalf("latest id = %d, want %d", latest.ID, second.ID)
	}
}
