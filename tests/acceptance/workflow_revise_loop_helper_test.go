// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// rlDir creates an isolated project dir with a .workflow/ folder.
func rlDir(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	if r, err := filepath.EvalSymlinks(d); err == nil {
		d = r
	}
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

// rlWrite writes body to dir/rel, creating parent directories.
func rlWrite(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// rlState seeds a workflow JSON for feature over order, sitting on current with
// every prior step done (each with a non-null CompletedAt) and the rest pending.
func rlState(t *testing.T, dir, feature string, order []string, current string) {
	t.Helper()
	steps := map[string]map[string]any{}
	done := true
	for _, s := range order {
		st := map[string]any{}
		switch {
		case s == current:
			st["status"], done = "in-progress", false
		case done:
			st["status"] = "done"
			ts := "2026-06-30T00:00:00Z"
			st["completedAt"] = ts
		default:
			st["status"] = "pending"
		}
		steps[s] = st
	}
	wf := map[string]any{
		"feature": feature, "currentStep": current, "stepOrder": order,
		"steps": steps, "startedAt": "2026-06-30T00:00:00Z",
	}
	b, _ := json.MarshalIndent(wf, "", "  ")
	rlWrite(t, dir, ".workflow/"+feature+".json", string(b))
}

// rlLoad reads back the workflow JSON as a generic map for assertions.
func rlLoad(t *testing.T, dir, feature string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".workflow", feature+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// rlStep returns the step sub-object from a loaded workflow map.
func rlStep(m map[string]any, step string) map[string]any {
	steps, _ := m["steps"].(map[string]any)
	s, _ := steps[step].(map[string]any)
	return s
}
