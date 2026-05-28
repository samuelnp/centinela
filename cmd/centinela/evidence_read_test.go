package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceReadField(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "feature-specialist"}); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceAppend(nil, []string{"alpha", "feature-specialist", "outputs", "docs/plans/alpha.md"}); err != nil {
		t.Fatal(err)
	}
	evidenceReadField = "outputs"
	t.Cleanup(func() { evidenceReadField = "" })
	out := captureStdout(t, func() {
		if err := runEvidenceRead(nil, []string{"alpha", "feature-specialist"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "docs/plans/alpha.md") {
		t.Fatalf("read --field outputs missing path: %q", out)
	}
}

func TestEvidenceReadMissingSuggestsInit(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceRead(nil, []string{"alpha", "qa-senior"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "evidence init alpha qa-senior") {
		t.Fatalf("hint missing: %v", err)
	}
}

func TestEvidenceReadFullDocument(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if err := runEvidenceRead(nil, []string{"alpha", "big-thinker"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"feature": "alpha"`) {
		t.Fatalf("full read missing feature: %q", out)
	}
}

// We rely on captureStdout defined in hook_cmd_test.go (same package).
// Touch os to ensure the import is used in case captureStdout's signature
// changes — kept light to stay G1-safe.
var _ = os.Stdout
