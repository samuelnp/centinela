package main

import (
	"os"
	"strings"
	"testing"
)

// seedFeatureBrief writes docs/features/<feature>.md so RequiredPlanInputs has a
// brief to snapshot (the init harness chdirs into a fresh tempdir).
func seedFeatureBrief(t *testing.T, feature string) {
	t.Helper()
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/"+feature+".md", []byte("brief"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceInitPreFillsPlanInputsForBigThinker(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "demo")
	seedFeatureBrief(t, "demo")
	if err := runEvidenceInit(nil, []string{"demo", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(".workflow/demo-big-thinker.json")
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"docs/plans/demo.md", "docs/features/demo.md"} {
		if !strings.Contains(s, want) {
			t.Fatalf("big-thinker init did not pre-fill %q: %s", want, s)
		}
	}
	if strings.Contains(s, "<FILL:") {
		t.Fatalf("fill marker leaked into JSON list field: %s", s)
	}
}

func TestEvidenceInitLeavesInputsEmptyForSeniorEngineer(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "demo")
	seedFeatureBrief(t, "demo")
	if err := runEvidenceInit(nil, []string{"demo", "senior-engineer"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(".workflow/demo-senior-engineer.json")
	if !strings.Contains(string(data), `"inputs": []`) {
		t.Fatalf("senior-engineer inputs not empty: %s", data)
	}
}
