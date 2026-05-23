package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

// markCmdDone seeds a done <feature>.json so FeatureStatus(feature) == "done".
// Shared across the cmd/centinela session + panel integration tests.
func markCmdDone(t *testing.T, feature string) {
	t.Helper()
	seedWF(t, feature, "done")
}

// seedWF saves a real <feature>.json workflow-state file at the given step,
// creating .workflow/ first so Save (which does not mkdir) succeeds.
func seedWF(t *testing.T, feature, step string) {
	t.Helper()
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatalf("mkdir workflow dir: %v", err)
	}
	wf := workflow.New(feature)
	wf.CurrentStep = step
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save %s: %v", feature, err)
	}
}

// runContext drives the real UserPromptSubmit hook end-to-end via stdout.
func runContext(t *testing.T) string {
	t.Helper()
	var out string
	withStdin(t, "{}", func() {
		out = captureStdout(t, func() {
			if err := runHookContext(nil, nil); err != nil {
				t.Fatalf("runHookContext returned error: %v", err)
			}
		})
	})
	return out
}

// Spec scenario 1: an evidence JSON in .workflow/ is NOT rendered active, while
// the genuine workflow-state file is — driven through loadActiveWorkflows + panel.
func TestRunHookContext_EvidenceJSONNotActive(t *testing.T) {
	chdirIntoTemp(t)
	seedWF(t, "alpha", "code")
	writeFile(t, ".workflow/alpha-qa-senior.json", `{"feature":"alpha","role":"qa-senior"}`)
	out := runContext(t)
	if !strings.Contains(out, "alpha") {
		t.Fatalf("expected active feature alpha, got:\n%s", out)
	}
	if strings.Contains(out, "qa-senior") {
		t.Fatalf("evidence JSON should not surface as an active workflow, got:\n%s", out)
	}
}

// Spec scenario 5: more than the cap of 5 active workflows shows at most 5 rows
// (most-recently-touched) plus a "+N more" hint.
func TestRunHookContext_CapShowsPlusNMore(t *testing.T) {
	chdirIntoTemp(t)
	base := time.Now().Add(-7 * time.Hour)
	for i := 0; i < 7; i++ {
		f := fmt.Sprintf("feat-%d", i)
		seedWF(t, f, "code") // "code" emits no per-feature reminder panels
		mt := base.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(workflow.FilePath(f), mt, mt); err != nil {
			t.Fatalf("chtimes %s: %v", f, err)
		}
	}
	out := runContext(t)
	if !strings.Contains(out, "+2 more") {
		t.Fatalf("expected '+2 more' hint with 7 active workflows, got:\n%s", out)
	}
	// The two oldest (feat-0, feat-1) are capped out; the newest (feat-6) shown.
	if strings.Contains(out, "feat-0") || strings.Contains(out, "feat-1") {
		t.Fatalf("oldest workflows should be capped out of the panel, got:\n%s", out)
	}
	if !strings.Contains(out, "feat-6") {
		t.Fatalf("most-recent workflow should be shown, got:\n%s", out)
	}
}
