package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestBuildStatusLineNoWorkflow(t *testing.T) {
	v := buildStatusLineView(nil)
	out := strings.Join(v.Primary, " ")
	if !strings.Contains(out, "WF:none") || !strings.Contains(out, "BLOCK:NO_WORKFLOW") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestBuildStatusLinePlanShowsWritePlan(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	mkdir(t, "docs/features")
	write(t, "docs/features/alpha.md", "x")
	wf := workflow.New("alpha")
	v := buildStatusLineView([]*workflow.Workflow{wf})
	out := strings.Join(v.Secondary, " ")
	if !strings.Contains(out, "NEXT:write-plan") {
		t.Fatalf("expected write-plan, got: %s", out)
	}
}

func TestBuildStatusLineTestsMissingEdgeCases(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	wf := workflow.New("alpha")
	wf.CurrentStep = "tests"
	v := buildStatusLineView([]*workflow.Workflow{wf})
	out := strings.Join(v.Secondary, " ")
	if !strings.Contains(out, "BLOCK:MISSING_EDGE_CASES") {
		t.Fatalf("expected missing edge cases block, got: %s", out)
	}
}

func withDir(t *testing.T, dir string) func() {
	t.Helper()
	o, _ := os.Getwd()
	os.Chdir(dir)                 //nolint:errcheck
	return func() { os.Chdir(o) } //nolint:errcheck
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
