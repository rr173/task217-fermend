package sampling

import (
	"errors"
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

func TestRegisterAndIngestSamples(t *testing.T) {
	s := openSamplingTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	sm := New(s)
	if _, err := sm.RegisterChannel(b.ID, "bad-do", model.KindDissolvedOxygen, model.UnitPH); !errors.Is(err, model.ErrUnitMismatch) {
		t.Fatalf("wrong unit error = %v, want ErrUnitMismatch", err)
	}
	ch, err := sm.RegisterChannel(b.ID, "do", model.KindDissolvedOxygen, model.UnitPercent)
	if err != nil {
		t.Fatal(err)
	}
	points := []model.SamplePoint{
		{Seq: 1, TUnix: 0, Value: 80},
		{Seq: 2, TUnix: 10, Value: 35},
	}
	if err := sm.IngestPoints(b.ID, ch.ID, points); err != nil {
		t.Fatal(err)
	}
	if err := sm.IngestPoints(b.ID, ch.ID, points); err != nil {
		t.Fatalf("duplicate batch should be idempotent, got %v", err)
	}
	got, err := sm.ListSamplesRange(ch.ID, 0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != 80 {
		t.Fatalf("range samples = %#v, want the first point only", got)
	}
	if err := sm.ExcludeChannel(ch.ID); err != nil {
		t.Fatal(err)
	}
	updated, err := sm.GetChannel(ch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != model.ChannelExcluded {
		t.Fatalf("channel status = %q, want excluded", updated.Status)
	}
}

func TestUnitAllowedCoversSupportedKinds(t *testing.T) {
	cases := []struct {
		kind model.ChannelKind
		unit model.ChannelUnit
	}{
		{model.KindPH, model.UnitPH},
		{model.KindFeed, model.UnitGramPerL},
		{model.KindTemperature, model.UnitCelsius},
		{model.KindAgitation, model.UnitRPM},
	}
	for _, tc := range cases {
		if !UnitAllowed(tc.kind, tc.unit) {
			t.Errorf("UnitAllowed(%q, %q) = false", tc.kind, tc.unit)
		}
	}
}

func openSamplingTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/sampling.db")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
