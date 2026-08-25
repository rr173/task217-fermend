package snapshot

import (
	"errors"
	"path/filepath"
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

// TestSnapshotEvidencePersists 验证证据在快照创建、发布后，以及服务重启
// （重开数据库）后，仍可通过列表、按版本、按 id、最新发布四条读取链路取回。
// 发布动作不得清空证据，列表与按版本读取不得返回空证据。
func TestSnapshotEvidencePersists(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "snapshot.db")

	// 首次会话：创建草稿、发布，并立即校验发布后证据仍在。
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	sn := New(s)
	encoded, err := EncodeEvidence(Evidence{EndpointT: 500, Verdict: "confirmed", DeviationSecs: 0, Notes: "verified"})
	if err != nil {
		t.Fatal(err)
	}
	draft, err := sn.CreateDraft(b.ID, 500, encoded)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Evidence != encoded {
		t.Fatalf("draft evidence = %q, want %q", draft.Evidence, encoded)
	}
	published, err := sn.Publish(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if published.Evidence != encoded {
		t.Fatalf("publish cleared evidence: got %q, want %q", published.Evidence, encoded)
	}
	// 列表与按版本读取在发布后同样必须返回证据，而非空串。
	gotByVersion, err := s.GetSnapshotByVersion(b.ID, draft.Version)
	if err != nil {
		t.Fatal(err)
	}
	if gotByVersion.Evidence != encoded {
		t.Fatalf("GetSnapshotByVersion evidence = %q, want %q", gotByVersion.Evidence, encoded)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	// 重启：重开同一数据库，证据应可被四条读取链路取回。
	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	sn2 := New(s2)

	gotByID, err := s2.GetSnapshot(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotByID.Evidence != encoded {
		t.Fatalf("after restart GetSnapshot evidence = %q, want %q", gotByID.Evidence, encoded)
	}
	gotByVersion2, err := s2.GetSnapshotByVersion(b.ID, draft.Version)
	if err != nil {
		t.Fatal(err)
	}
	if gotByVersion2.Evidence != encoded {
		t.Fatalf("after restart GetSnapshotByVersion evidence = %q, want %q", gotByVersion2.Evidence, encoded)
	}
	list, err := sn2.ListSnapshots(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Evidence != encoded {
		t.Fatalf("after restart ListSnapshots evidence = %v", list)
	}
	latest, err := sn2.LatestPublished(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Evidence != encoded {
		t.Fatalf("after restart LatestPublished evidence = %q, want %q", latest.Evidence, encoded)
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
