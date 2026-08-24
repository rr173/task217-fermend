package calibration

import (
	"testing"

	"task217-fermend/internal/model"
)

func TestLagPrimitives(t *testing.T) {
	points := []*model.SamplePoint{
		{TUnix: 10, Value: 4},
		{TUnix: 20, Value: 1},
		{TUnix: 30, Value: 3},
	}
	if got, value, ok := FindMinTime(points); !ok || got != 20 || value != 1 {
		t.Fatalf("FindMinTime = (%d, %v, %v), want (20, 1, true)", got, value, ok)
	}
	if got, ok := FindMaxSlopeTime(points); !ok || got != 20 {
		t.Fatalf("FindMaxSlopeTime = (%d, %v), want (20, true)", got, ok)
	}
	if got := EstimateLag(500, 530); got != 30 {
		t.Fatalf("EstimateLag = %v, want 30", got)
	}
}

func TestGapDetectReturnsMissingWindows(t *testing.T) {
	points := []*model.SamplePoint{
		{TUnix: 0},
		{TUnix: 10},
		{TUnix: 30},
	}
	gaps := GapDetect(points, 15)
	if len(gaps) != 1 || gaps[0].StartUnix != 10 || gaps[0].EndUnix != 30 || gaps[0].Status != model.SegmentGap {
		t.Fatalf("gaps = %#v, want one gap [10,30]", gaps)
	}
}
