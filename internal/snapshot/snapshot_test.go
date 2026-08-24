package snapshot

import (
	"errors"
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

func TestSnapshotLifecycle(t *testing.T) {
	s := openSnapshotTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	sn := New(s)
	evidence := Evidence{EndpointT: 500, Verdict: "confirmed", DeviationSecs: 0, Notes: "verified"}
	encoded, err := EncodeEvidence(evidence)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeEvidence(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.EndpointT != 500 || decoded.Verdict != "confirmed" {
		t.Fatalf("decoded evidence = %#v", decoded)
	}
	draft, err := sn.CreateDraft(b.ID, 500, encoded)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Version != 1 || draft.Status != model.SnapshotDraft {
		t.Fatalf("draft = %#v", draft)
	}
	published, err := sn.Publish(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != model.SnapshotPublished {
		t.Fatalf("published status = %q", published.Status)
	}
	if _, err := sn.Publish(draft.ID); !errors.Is(err, model.ErrSnapshotSealed) {
		t.Fatalf("republish error = %v, want ErrSnapshotSealed", err)
	}
	latest, err := sn.LatestPublished(b.ID)
	if err != nil || latest.ID != draft.ID {
		t.Fatalf("LatestPublished = %#v, %v", latest, err)
	}
}

func openSnapshotTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/snapshot.db")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
