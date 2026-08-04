// Acceptance: specs/hook-context-panel-diet.feature
//
// Shared binary + fixture helpers for the hook-context-panel-diet acceptance
// suite. Everything is local: the binary is built from ./cmd/centinela into a
// temp dir, and every fixture lives in t.TempDir() — no network, no shared
// install path, and never the real repository's .workflow/ state.
package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var pdBinOnce sync.Once
var pdBin string

func pdBuildBin(t *testing.T) string {
	t.Helper()
	pdBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-pd-bin")
		if err != nil {
			t.Fatal(err)
		}
		pdBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", pdBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return pdBin
}

// pdRepo seeds a throwaway project directory with .workflow/.
func pdRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pdWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func pdWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// pdCanonical is the canonical five-step order; archetype fixtures pass their
// own subset so the rendered ladder differs per workflow.
var pdCanonical = []string{"plan", "code", "tests", "validate", "docs"}

func pdWorkflow(t *testing.T, dir, feature, step string, order []string) {
	t.Helper()
	quoted := make([]string, len(order))
	for i, s := range order {
		quoted[i] = `"` + s + `"`
	}
	body := `{"feature":"` + feature + `","currentStep":"` + step + `",` +
		`"stepOrder":[` + strings.Join(quoted, ",") + `],"steps":{}}`
	pdWrite(t, filepath.Join(dir, ".workflow", feature+".json"), body)
}

// pdPlanArtifacts writes the plan-step artifacts that make ValidateArtifacts
// pass, which is what arms the REVIEW REQUIRED nudge panel.
func pdPlanArtifacts(t *testing.T, dir, feature string) {
	t.Helper()
	pdWrite(t, filepath.Join(dir, "docs", "features", feature+".md"), "# "+feature+"\n")
	pdWrite(t, filepath.Join(dir, "docs", "plans", feature+".md"), "# Plan\n")
	pdWrite(t, filepath.Join(dir, "specs", feature+".feature"), "Feature: "+feature+"\n")
}

// pdHookContext runs the UserPromptSubmit hook and returns its combined output.
func pdHookContext(t *testing.T, bin, dir string) string {
	t.Helper()
	c := exec.Command(bin, "hook", "context")
	c.Dir = dir
	c.Stdin = strings.NewReader(`{"session_id":"pd-1"}`)
	out, err := c.CombinedOutput()
	if _, ok := err.(*exec.ExitError); err != nil && !ok {
		t.Fatalf("run hook context: %v", err)
	}
	return string(out)
}

// pdPaddedLines returns every line still carrying lipgloss right-padding.
func pdPaddedLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != strings.TrimRight(line, " \t") {
			out = append(out, line)
		}
	}
	return out
}

func pdRequireUnpadded(t *testing.T, label, out string) {
	t.Helper()
	if bad := pdPaddedLines(out); len(bad) > 0 {
		t.Fatalf("%s emitted %d padded line(s), first %q\nfull output:\n%s", label, len(bad), bad[0], out)
	}
}

// pdRunGuard runs the colocated size guard, optionally injecting synthetic
// padding through the guard's test-only env knob, and returns output + code.
func pdRunGuard(t *testing.T, run string, padBytes string) (string, int) {
	t.Helper()
	c := exec.Command("go", "test", "./internal/ui/", "-run", run, "-count=1", "-v")
	c.Dir = repoRoot(t)
	c.Env = os.Environ()
	if padBytes != "" {
		c.Env = append(c.Env, "CENTINELA_PANEL_DIET_PAD="+padBytes)
	}
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run size guard: %v", err)
	}
	return string(out), code
}

// pdReadGuardSource returns the colocated size-guard source for structural
// assertions about what it is allowed to read.
func pdReadGuardSource(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "internal", "ui", "panel_budget_test.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read size guard source: %v", err)
	}
	return string(data)
}
