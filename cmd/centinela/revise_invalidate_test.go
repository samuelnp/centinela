package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestInvalidateDownstreamCountsAndDedups(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	for _, r := range []string{"validation-specialist", "gatekeeper"} {
		os.WriteFile(".workflow/f-"+r+".json", []byte("x"), 0o644) //nolint:errcheck
	}
	os.WriteFile(".workflow/f-qa-senior.json", []byte("x"), 0o644) //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("x"), 0o644)  //nolint:errcheck
	// "tests" twice exercises both role + artifact dedup; production-readiness
	// is absent so it removes nothing and does not count.
	count, err := invalidateDownstream("f", []string{"tests", "validate", "tests"})
	if err != nil {
		t.Fatalf("invalidateDownstream: %v", err)
	}
	// removed: qa-senior, edge-cases.md, validation-specialist, gatekeeper = 4.
	if count != 4 {
		t.Fatalf("count = %d, want 4", count)
	}
}

func TestInvalidateDownstreamErrorSurfaces(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	// A non-empty dir at an evidence path makes os.Remove fail non-idempotently.
	os.MkdirAll(".workflow/f-qa-senior.json/child", 0o755) //nolint:errcheck
	if _, err := invalidateDownstream("f", []string{"tests"}); err == nil {
		t.Fatal("expected removal error to surface")
	}
}
