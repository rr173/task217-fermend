package stage

import (
	"errors"
	"testing"

	"task217-fermend/internal/model"
	"task217-fermend/internal/store"
)

func TestAddStageRejectsInvalidOrder(t *testing.T) {
	s := openStageTestStore(t)
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch-1", "strain-1", "BR-1"))
	if err != nil {
		t.Fatal(err)
	}
	st := New(s)
	if _, err := st.AddStage(b.ID, "lag", 100, 200); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddStage(b.ID, "bad-end", 100, 100); !errors.Is(err, model.ErrStageOrder) {
		t.Fatalf("equal stage bounds error = %v, want ErrStageOrder", err)
	}
	if _, err := st.AddStage(b.ID, "bad-start", 50, 150); !errors.Is(err, model.ErrStageOrder) {
		t.Fatalf("reverse stage start error = %v, want ErrStageOrder", err)
	}
	stages, err := st.ListStages(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 1 || stages[0].Seq != 1 {
		t.Fatalf("stored stages = %#v, want one stage with seq 1", stages)
	}
}

func TestLocateUsesHalfOpenWindows(t *testing.T) {
	windows := BuildWindows([]*model.Stage{
		{Name: "lag", StartUnix: 0, EndUnix: 10},
		{Name: "log", StartUnix: 10, EndUnix: 0},
	})
	if got := Locate(windows, 10); got != "log" {
		t.Fatalf("Locate at boundary = %q, want log", got)
	}
	if windows[0].Covering(10) {
		t.Fatal("first half-open window must not cover its end")
	}
}

func openStageTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/stage.db")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
