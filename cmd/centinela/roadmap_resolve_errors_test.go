package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// The production seam must really shell out to git; a stub can never prove it.
func TestGitStageRunnerAndCommandEntryPoint(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	body := doc(`{"name":"f","deferredAt":"2026-01-01T00:00:00Z"}`)
	resolveRepo(t, body)
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"add", "-A"}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	if out, err := gitStageRunner(".", "ls-files", "--unmerged", "--", ".workflow/roadmap.json"); err != nil ||
		strings.TrimSpace(out) != "" {
		t.Fatalf("a clean index has no unmerged entry: %q %v", out, err)
	}
	out := captureStdout(t, func() {
		if err := runRoadmapResolve(nil, nil); err != nil {
			t.Errorf("no-op resolve must exit 0: %v", err)
		}
	})
	if !strings.Contains(out, "Nothing to resolve") {
		t.Fatalf("output = %q", out)
	}
}

// A staging failure is surfaced, not swallowed: the operator must know the
// conflict is still unresolved as far as git is concerned.
func TestResolveSurfacesAStagingFailure(t *testing.T) {
	resolveRepo(t, markers)
	run := roadmap.StageRunner(func(_ string, args ...string) (string, error) {
		switch args[0] {
		case "ls-files":
			return "unmerged\n", nil
		case "show":
			return doc(`{"name":"a","deferredAt":"2026-01-01T00:00:00Z"}`), nil
		}
		return "index locked", errors.New("exit status 1")
	})
	err := resolveRoadmapState(".", run)
	if err == nil || !strings.Contains(err.Error(), "staging resolved roadmap state") {
		t.Fatalf("want a staging error, got %v", err)
	}
}

// A regeneration failure (ROADMAP.md is a directory) must surface too.
func TestResolveSurfacesARegenerationFailure(t *testing.T) {
	resolveRepo(t, doc(`{"name":"a","deferredAt":"2026-01-01T00:00:00Z"}`))
	if err := os.Mkdir("ROADMAP.md", 0o755); err != nil {
		t.Fatal(err)
	}
	var staged []string
	run := resolveStub(map[string]string{"ROADMAP.md": "unmerged\n"}, nil, &staged)
	if err := resolveRoadmapState(".", run); err == nil {
		t.Fatal("a failed regeneration must surface")
	}
	if len(staged) != 0 {
		t.Fatalf("nothing may be staged after a failure: %v", staged)
	}
}

// A conflicted path whose stages cannot be read is refused, not merged blind.
func TestResolveSurfacesAnUnreadableStage(t *testing.T) {
	resolveRepo(t, markers)
	var staged []string
	run := resolveStub(map[string]string{".workflow/roadmap.json": "unmerged\n"}, nil, &staged)
	if err := resolveRoadmapState(".", run); err == nil ||
		!strings.Contains(err.Error(), "neither side") {
		t.Fatalf("want a contentless-conflict refusal, got %v", err)
	}
	body, _ := os.ReadFile(".workflow/roadmap.json") //nolint:errcheck
	if string(body) != markers {
		t.Fatal("the markers must be untouched")
	}
}
