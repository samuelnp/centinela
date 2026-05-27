package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// chdirWorkflow moves into a fresh temp dir with .workflow/ so roadmap.Load and
// FeatureStatus operate on real on-disk fixtures (roadmap.json + workflow files).
func chdirWorkflow(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeRoadmap(t *testing.T, body string) {
	t.Helper()
	if err := os.WriteFile(roadmap.RoadmapFile, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func seedDone(t *testing.T, name string) {
	t.Helper()
	if err := workflow.Save(&workflow.Workflow{Feature: name, CurrentStep: "done", Steps: map[string]workflow.StepState{}}); err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
}

// rehydrate mirrors cmd hook_session: load the real roadmap, derive the ready set
// + incomplete flag from disk, and render the cross-layer rehydration payload.
func rehydrate(t *testing.T) string {
	t.Helper()
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("roadmap.Load: %v", err)
	}
	ready := roadmap.ReadySet(r)
	planned, inProgress, _ := r.Summary()
	return ui.RenderSessionRehydration(r, ready, planned > 0 || inProgress > 0)
}

// End-to-end roadmap render: with feature-a done, the no-dep frontier feature and
// the now-unblocked dependent both annotate and the blocked feature shows 🔒 + dep.
func TestIntegration_RoadmapRenderAnnotatesAndListsFrontier(t *testing.T) {
	chdirWorkflow(t)
	writeRoadmap(t, `{"phases":[{"name":"P","features":[
		{"name":"feature-a"},
		{"name":"feature-b","dependsOn":["feature-a"]},
		{"name":"feature-c"},
		{"name":"feature-x","dependsOn":["feature-c"]}]}]}`)
	seedDone(t, "feature-a")
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	render := ui.RenderRoadmap(r)
	if !strings.Contains(lineWith(render, "feature-b"), "🔓") {
		t.Fatalf("feature-b should be ready (🔓) after dep done:\n%s", render)
	}
	if !strings.Contains(lineWith(render, "feature-x"), "🔒") || !strings.Contains(render, "blocked-by: feature-c") {
		t.Fatalf("feature-x should be 🔒 blocked-by feature-c:\n%s", render)
	}
	out := rehydrate(t)
	if !strings.Contains(out, "Ready to start now:") {
		t.Fatalf("expected plural frontier header:\n%s", out)
	}
	for _, name := range []string{"feature-b", "feature-c"} {
		if !strings.Contains(out, name) {
			t.Fatalf("frontier should list %q:\n%s", name, out)
		}
	}
}

// Empty frontier with planned work remaining must NOT look complete and must
// surface the blocking reason instead of any ready feature.
func TestIntegration_EmptyFrontierBlockedNotComplete(t *testing.T) {
	chdirWorkflow(t)
	writeRoadmap(t, `{"phases":[{"name":"P","features":[
		{"name":"base"},{"name":"leaf","dependsOn":["base"]}]}]}`)
	// base is planned (no workflow file) → leaf is blocked, base itself ready.
	// Make base in-progress so NOTHING is ready yet work remains.
	if err := workflow.Save(&workflow.Workflow{Feature: "base", CurrentStep: "code", Steps: map[string]workflow.StepState{}}); err != nil {
		t.Fatal(err)
	}
	out := rehydrate(t)
	if strings.Contains(out, "Roadmap complete") {
		t.Fatalf("blocked-but-incomplete must not show roadmap-complete:\n%s", out)
	}
	if !strings.Contains(out, "blocked by unmet dependencies") && !strings.Contains(out, "in-progress or blocked") {
		t.Fatalf("must reference the blocking reason:\n%s", out)
	}
}

// All features done → roadmap-complete message and no ready features listed.
func TestIntegration_AllDoneShowsComplete(t *testing.T) {
	chdirWorkflow(t)
	writeRoadmap(t, `{"phases":[{"name":"P","features":[
		{"name":"one"},{"name":"two","dependsOn":["one"]}]}]}`)
	seedDone(t, "one")
	seedDone(t, "two")
	out := rehydrate(t)
	if !strings.Contains(out, "Roadmap complete") {
		t.Fatalf("all-done should show roadmap-complete:\n%s", out)
	}
	if strings.Contains(out, "Ready to start now:") {
		t.Fatalf("all-done must not list a ready frontier:\n%s", out)
	}
}

func lineWith(s, sub string) string {
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, sub) {
			return ln
		}
	}
	return ""
}
