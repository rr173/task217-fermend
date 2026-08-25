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

func TestSnapshotSupersede(t *testing.T) {
	s := openSnapshotTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	sn := New(s)
	evidence, err := EncodeEvidence(Evidence{EndpointT: 500, Verdict: "confirmed", Notes: "v1"})
	if err != nil {
		t.Fatal(err)
	}

	// v1 草稿 → 发布
	old, err := sn.CreateDraft(b.ID, 500, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sn.Publish(old.ID); err != nil {
		t.Fatal(err)
	}

	// v2 新草稿
	newSnap, err := sn.CreateDraft(b.ID, 600, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if newSnap.Version != 2 {
		t.Fatalf("new version = %d, want 2", newSnap.Version)
	}

	// 用 v2 替代 v1：旧 → superseded，新 → published
	pub, err := sn.Supersede(old.ID, newSnap.ID)
	if err != nil {
		t.Fatalf("Supersede error = %v", err)
	}
	if pub.ID != newSnap.ID || pub.Status != model.SnapshotPublished {
		t.Fatalf("Supersede result = %#v, want new snapshot published", pub)
	}

	// 旧快照必须进入 superseded
	list, err := sn.ListSnapshots(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var oldAfter *model.DiagnosisSnapshot
	for _, x := range list {
		if x.ID == old.ID {
			oldAfter = x
		}
	}
	if oldAfter == nil || oldAfter.Status != model.SnapshotSuperseded {
		t.Fatalf("old after supersede = %#v, want superseded", oldAfter)
	}

	// 最新发布快照必须是新版本
	latest, err := sn.LatestPublished(b.ID)
	if err != nil || latest.ID != newSnap.ID || latest.Version != 2 {
		t.Fatalf("LatestPublished = %#v, %v", latest, err)
	}

	// 已封存（superseded）快照不可再次被替代
	newer, err := sn.CreateDraft(b.ID, 700, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sn.Supersede(old.ID, newer.ID); !errors.Is(err, model.ErrSnapshotSealed) {
		t.Fatalf("resupersede superseded-old error = %v, want ErrSnapshotSealed", err)
	}
	// 旧快照必须是 published：草稿不能作为被替代的旧快照
	newer2, err := sn.CreateDraft(b.ID, 800, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sn.Supersede(newer.ID, newer2.ID); !errors.Is(err, model.ErrSnapshotSealed) {
		t.Fatalf("supersede draft-old error = %v, want ErrSnapshotSealed", err)
	}
	// 新快照必须是 draft：published 不能作为新快照
	if _, err := sn.Supersede(newSnap.ID, old.ID); !errors.Is(err, model.ErrSnapshotSealed) {
		t.Fatalf("supersede published-new error = %v, want ErrSnapshotSealed", err)
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
