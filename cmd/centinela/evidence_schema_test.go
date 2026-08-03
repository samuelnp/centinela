package main

import (
	"strings"
	"testing"
)

func TestEvidenceSchemaEmitsSkeleton(t *testing.T) {
	// The skeleton derives its feature from the CWD, so the no-feature branch
	// must be asserted from a directory that resolves nothing: a temp dir (no
	// `.worktrees/<feature>` segment) with an empty .workflow. Without this the
	// test's outcome depends on where the repo happens to be checked out — it
	// fails outright when the checkout is itself inside a `.worktrees/` path.
	chdirEvidenceTemp(t)
	out := captureStdout(t, func() {
		if err := runEvidenceSchema(nil, []string{"big-thinker"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"feature": "<feature-slug>"`) {
		t.Fatalf("schema missing placeholder feature: %q", out)
	}
	if !strings.Contains(out, `"role": "big-thinker"`) {
		t.Fatalf("schema missing role: %q", out)
	}
}

func TestEvidenceSchemaRejectsUnknownRole(t *testing.T) {
	err := runEvidenceSchema(nil, []string{"bogus"})
	if err == nil || !strings.Contains(err.Error(), "unknown role") {
		t.Fatalf("expected unknown role, got %v", err)
	}
}
