package acceptance_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// Acceptance: specs/team-dashboard.feature
//
// Drives the built binary `centinela dashboard` (+ --json) in a temp dir seeded
// with a .workflow/<f>.json and roadmap.json. LOCAL only — no git push, no
// network. The owner column resolves via the real (local) git seam; a non-repo
// temp dir makes git log fail → "unknown", which is an accepted advisory value.

func dashboardDir(t *testing.T, feature, step string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, ".workflow/"+feature+".json",
		`{"feature":"`+feature+`","currentStep":"`+step+`","steps":{}}`)
	writeFile(t, dir, ".workflow/roadmap.json",
		`{"phases":[{"name":"Q1","features":[{"name":"f1"},{"name":"f2"}]}]}`)
	return dir
}

// Scenario: Dashboard prints three panels from current on-disk state and exits 0
func TestDashboardPrintsThreePanels(t *testing.T) {
	out, code := runCent(t, buildCent(t), dashboardDir(t, "alpha", "code"), "dashboard")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	for _, want := range []string{"In-flight features", "Roadmap burn-down", "Gate health", "alpha"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

// Scenario: All three sources missing yields three honest empty panels and exits 0
func TestDashboardEmptyStates(t *testing.T) {
	out, code := runCent(t, buildCent(t), t.TempDir(), "dashboard")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	for _, want := range []string{"no active features", "no roadmap", "no gate failures recorded"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing empty state %q:\n%s", want, out)
		}
	}
}

// Scenario: --json emits full Dashboard struct as indented JSON and exits 0
func TestDashboardJSONKeys(t *testing.T) {
	out, code := runCent(t, buildCent(t), dashboardDir(t, "alpha", "code"), "dashboard", "--json")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("json must be ANSI-free:\n%s", out)
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
