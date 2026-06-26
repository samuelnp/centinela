package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runDashboardCapture chdirs into dir, stubs the git owner seam, runs the
// command, and returns stdout. Restores cwd, flag, and seam afterwards.
func runDashboardCapture(t *testing.T, dir string, asJSON bool) string {
	t.Helper()
	orig, _ := os.Getwd()
	prevOwner := gitOwner
	t.Cleanup(func() { _ = os.Chdir(orig); gitOwner = prevOwner; dashboardJSON = false })
	gitOwner = func(string, string) string { return "tester" }
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	dashboardJSON = asJSON
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	err := runDashboard(nil, nil)
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("runDashboard: %v", err)
	}
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func seedWorkflow(t *testing.T, dir, feature, step string) {
	t.Helper()
	wd := filepath.Join(dir, ".workflow")
	if err := os.MkdirAll(wd, 0o755); err != nil {
		t.Fatal(err)
	}
	js := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(filepath.Join(wd, feature+".json"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDashboard_HappyPanels(t *testing.T) {
	dir := t.TempDir()
	seedWorkflow(t, dir, "alpha", "code")
	out := runDashboardCapture(t, dir, false)
	for _, want := range []string{"In-flight features", "Roadmap burn-down", "Gate health", "alpha", "tester"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRunDashboard_EmptyStates(t *testing.T) {
	out := runDashboardCapture(t, t.TempDir(), false)
	for _, want := range []string{"no active features", "no roadmap", "no gate failures recorded"} {
		if !strings.Contains(out, want) {
			t.Fatalf("empty state missing %q:\n%s", want, out)
		}
	}
}

func TestRunDashboard_JSONShape(t *testing.T) {
	dir := t.TempDir()
	seedWorkflow(t, dir, "alpha", "code")
	out := runDashboardCapture(t, dir, true)
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("json must be ANSI-free: %q", out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	for _, f := range []string{"Features", "Roadmap", "Gates"} {
		if _, ok := m[f]; !ok {
			t.Fatalf("missing top-level field %q: %v", f, m)
		}
	}
}
