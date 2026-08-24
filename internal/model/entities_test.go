package model

import "testing"

func TestBatchTransitionRules(t *testing.T) {
	b := NewBatch("batch-1", "strain-1", "BR-1")
	if b.Status != BatchPreparing {
		t.Fatalf("new batch status = %q, want %q", b.Status, BatchPreparing)
	}
	if !b.CanTransition(BatchRunning) {
		t.Fatal("preparing batch should transition to running")
	}
	if b.CanTransition(BatchConfirmed) {
		t.Fatal("preparing batch must not skip to confirmed")
	}
	if !b.CanTransition(BatchPreparing) {
		t.Fatal("same-state transition should be idempotently allowed")
	}
}
