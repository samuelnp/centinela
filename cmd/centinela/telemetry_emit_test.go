package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/verify"
)

// TestEmitGateFailuresRecordsOnlyFails records one event per failing gate and
// skips Pass results (the previously-uncovered Fail branch).
func TestEmitGateFailuresRecordsOnlyFails(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	results := []gates.Result{
		{Name: "G1", Status: gates.Fail, Message: "too big"},
		{Name: "G2", Status: gates.Pass},
		{Name: "G3", Status: gates.Fail, Message: "bad import"},
	}
	emitGateFailures(cfg, results, "claude-opus")
	events, rerr := telemetry.ReadDefault()
	if rerr != nil {
		t.Fatal(rerr)
	}
	var fails int
	for _, e := range events {
		if e.Type == telemetry.TypeGateFailure {
			fails++
		}
	}
	if fails != 2 {
		t.Fatalf("expected 2 gate-failure events (only Fail results), got %d", fails)
	}
}

// TestToCheckRefsEmptyReturnsNil covers the len==0 short-circuit so an empty
// failure set never allocates a slice.
func TestToCheckRefsEmptyReturnsNil(t *testing.T) {
	if refs := toCheckRefs(nil); refs != nil {
		t.Fatalf("empty checks should map to nil, got %v", refs)
	}
}

// TestToCheckRefsCopiesFields proves the verify.Check → telemetry.CheckRef copy.
func TestToCheckRefsCopiesFields(t *testing.T) {
	in := []verify.Check{{Claim: "c", Role: "r", Status: verify.StatusFail, Detail: "d"}}
	refs := toCheckRefs(in)
	if len(refs) != 1 || refs[0].Claim != "c" || refs[0].Role != "r" || refs[0].Detail != "d" {
		t.Fatalf("CheckRef copy wrong: %+v", refs)
	}
}
