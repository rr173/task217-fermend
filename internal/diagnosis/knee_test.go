package diagnosis

import (
	"testing"

	"task217-fermend/internal/model"
)

func TestKneeDetectors(t *testing.T) {
	do := []*model.SamplePoint{{TUnix: 0, Value: 80}, {TUnix: 10, Value: 30}, {TUnix: 20, Value: 60}}
	if got, _, ok := DoReboundTime(do); !ok || got != 10 {
		t.Fatalf("DoReboundTime = (%d, %v), want (10, true)", got, ok)
	}
	ph := []*model.SamplePoint{{TUnix: 0, Value: 7}, {TUnix: 10, Value: 6.9}, {TUnix: 20, Value: 8.5}}
	if got, _, ok := PhKneeTime(ph); !ok || got != 20 {
		t.Fatalf("PhKneeTime = (%d, %v), want (20, true)", got, ok)
	}
	feed := []*model.SamplePoint{{TUnix: 0, Value: 0}, {TUnix: 10, Value: 1}, {TUnix: 20, Value: 5}, {TUnix: 30, Value: 15}}
	if got, _, ok := FeedKneeTime(feed); !ok || got != 30 {
		t.Fatalf("FeedKneeTime = (%d, %v), want (30, true)", got, ok)
	}
}

func TestMedianAndComparison(t *testing.T) {
	if got := MedianTime([]int64{530, 500, 500}); got != 500 {
		t.Fatalf("MedianTime = %d, want 500", got)
	}
	if got := MedianTime([]int64{500, 530}); got != 530 {
		t.Fatalf("even MedianTime = %d, want upper middle 530", got)
	}
	d := &Diagnosis{ToleranceSecs: 30}
	if got := d.Compare(500, 530); !got.WithinTolerance || got.Verdict != "confirmed" || got.DeviationSecs != 30 {
		t.Fatalf("Compare at tolerance = %#v, want confirmed with deviation 30", got)
	}
	if got := d.Compare(500, 531); got.WithinTolerance || got.Verdict != "conflict" {
		t.Fatalf("Compare beyond tolerance = %#v, want conflict", got)
	}
}
