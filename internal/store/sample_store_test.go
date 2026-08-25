package store

import (
	"testing"
	"time"

	"task217-fermend/internal/model"
)

func openSampleTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir() + "/samples.db")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestInsertSamplesTxKeepsFirstValueOnDuplicateSeq(t *testing.T) {
	s := openSampleTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	ch, err := s.CreateChannel(&model.Channel{
		BatchID: b.ID, Name: "do", Kind: model.KindDissolvedOxygen,
		Unit: model.UnitPercent, Status: model.ChannelActive, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	first := []*model.SamplePoint{
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 1, TUnix: 100, Value: 80, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 2, TUnix: 200, Value: 35, CreatedAt: time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC)},
	}
	if err := s.InsertSamplesTx(first); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	// 第二次重复提交同一 (batch_id, channel_id, seq)，但带不同的时间与值——必须幂等、保留首值。
	dup := []*model.SamplePoint{
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 1, TUnix: 999, Value: 7, CreatedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)},
		{BatchID: b.ID, ChannelID: ch.ID, Seq: 2, TUnix: 888, Value: 9, CreatedAt: time.Date(2030, 1, 1, 0, 0, 1, 0, time.UTC)},
	}
	if err := s.InsertSamplesTx(dup); err != nil {
		t.Fatalf("duplicate ingest should be idempotent, got %v", err)
	}
	// 任一 (batch, channel, seq) 只能保留第一次的时间与值。
	got, err := s.ListSamplesRange(ch.ID, 0, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique samples, got %d (duplicates not deduped)", len(got))
	}
	if got[0].Seq != 1 || got[0].TUnix != 100 || got[0].Value != 80 {
		t.Fatalf("first sample overwritten = %#v, want (seq=1 t=100 v=80)", got[0])
	}
	if got[1].Seq != 2 || got[1].TUnix != 200 || got[1].Value != 35 {
		t.Fatalf("second sample overwritten = %#v, want (seq=2 t=200 v=35)", got[1])
	}
}

func TestInsertSampleReturnsDuplicate(t *testing.T) {
	s := openSampleTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	ch, err := s.CreateChannel(&model.Channel{
		BatchID: b.ID, Name: "do", Kind: model.KindDissolvedOxygen,
		Unit: model.UnitPercent, Status: model.ChannelActive, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	p := &model.SamplePoint{BatchID: b.ID, ChannelID: ch.ID, Seq: 1, TUnix: 100, Value: 80, CreatedAt: time.Now().UTC()}
	if err := s.InsertSample(p); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := s.InsertSample(&model.SamplePoint{BatchID: b.ID, ChannelID: ch.ID, Seq: 1, TUnix: 999, Value: 7, CreatedAt: time.Now().UTC()}); err == nil {
		t.Fatal("duplicate insert should return ErrDuplicate")
	}
	got, err := s.ListSamplesRange(ch.ID, 0, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].TUnix != 100 || got[0].Value != 80 {
		t.Fatalf("first value not preserved = %#v", got)
	}
}

func TestStoreDoesNotPanicWhenUninitialized(t *testing.T) {
	var s *Store // nil store, as if store.Open failed and caller ignored err
	if _, err := s.CreateBatch(model.NewBatch("x", "y", "z")); err == nil {
		t.Fatal("expected ErrStoreUnavailable, got nil")
	}
	var zs *Store
	if err := zs.InsertSamplesTx(nil); err == nil {
		t.Fatal("expected ErrStoreUnavailable, got nil")
	}
	if zs.DB() != nil {
		t.Fatal("nil Store.DB should return nil, not panic")
	}
	if zs.Ready() {
		t.Fatal("nil Store.Ready should be false")
	}
}

func TestSamplesTableMigrationSucceeds(t *testing.T) {
	s := openSampleTestStore(t)
	defer s.Close()
	// migration ran in Open; verify the samples table + unique constraint exist.
	var name string
	if err := s.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='samples'`).Scan(&name); err != nil {
		t.Fatalf("samples table missing: %v", err)
	}
	// second migration must be idempotent (no syntax error).
	if err := s.migrate(); err != nil {
		t.Fatalf("re-run migrate should be idempotent, got %v", err)
	}
}
