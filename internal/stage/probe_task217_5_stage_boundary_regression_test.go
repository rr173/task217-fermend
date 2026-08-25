package stage_test

import (
	"task217-fermend/internal/model"
	"task217-fermend/internal/stage"
	"task217-fermend/internal/store"
	"testing"
)

func TestAdjacentStagesShareBoundary(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/stage.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b, err := s.CreateBatch(model.NewBatch("batch", "strain", "BR"))
	if err != nil {
		t.Fatal(err)
	}
	st := stage.New(s)
	if _, err := st.AddStage(b.ID, "lag", 0, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AddStage(b.ID, "log", 100, 200); err != nil {
		t.Fatalf("adjacent stage rejected: %v", err)
	}
	windows := stage.BuildWindows([]*model.Stage{{StartUnix: 0, EndUnix: 100}, {Name: "log", StartUnix: 100, EndUnix: 200}})
	if got := stage.Locate(windows, 100); got != "log" {
		t.Fatalf("boundary located in %q", got)
	}
}
