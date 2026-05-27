package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// runRoadmapReady loads the roadmap and prints the ready frontier; with feature-a
// done, the no-dep feature-c and the unblocked dependent feature-b are ready.
func TestRunRoadmapReady_PrintsReadyFrontier(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"P","features":[{"name":"feature-a"},`+
			`{"name":"feature-b","dependsOn":["feature-a"]},{"name":"feature-c"}]}]}`)
	markCmdDone(t, "feature-a")
	out := captureStdout(t, func() {
		if err := runRoadmapReady(nil, nil); err != nil {
			t.Fatalf("runRoadmapReady: %v", err)
		}
	})
	for _, name := range []string{"feature-b", "feature-c"} {
		if !strings.Contains(out, name) {
			t.Fatalf("ready output should list %q:\n%s", name, out)
		}
	}
	if strings.Contains(out, "feature-a") {
		t.Fatalf("done feature-a should not be in the ready frontier:\n%s", out)
	}
}

// Empty frontier prints a non-empty empty-state and no feature names.
func TestRunRoadmapReady_EmptyState(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"P","features":[{"name":"a"},{"name":"b","dependsOn":["a"]}]}]}`)
	seedWF(t, "a", "code")
	out := captureStdout(t, func() {
		if err := runRoadmapReady(nil, nil); err != nil {
			t.Fatalf("runRoadmapReady: %v", err)
		}
	})
	if strings.TrimSpace(out) == "" {
		t.Fatalf("empty frontier must print a non-empty empty-state line")
	}
	if strings.Contains(out, "🔓 a") || strings.Contains(out, "🔓 b") {
		t.Fatalf("empty-state must not list feature names:\n%s", out)
	}
}

// Missing/invalid roadmap surfaces a setup error rather than panicking.
func TestRunRoadmapReady_MissingRoadmapErrors(t *testing.T) {
	chdirIntoTemp(t)
	err := runRoadmapReady(nil, nil)
	if err == nil || !strings.Contains(err.Error(), roadmap.RoadmapFile) {
		t.Fatalf("missing roadmap should error naming %s, got %v", roadmap.RoadmapFile, err)
	}
}
