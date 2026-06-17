package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rdsToml enables only the roadmap_drift gate so validate output isolates it.
func rdsToml(severity string) string {
	return "[gates]\nfile_size = false\n\n[gates.roadmap_drift]\nenabled = true\nseverity = \"" +
		severity + "\"\n"
}

// rdsDisabledToml ships the gate disabled.
const rdsDisabledToml = "[gates]\nfile_size = false\n\n[gates.roadmap_drift]\nenabled = false\n"

// rdsDir builds a temp project with roadmap.json and a centinela.toml.
func rdsDir(t *testing.T, roadmapJSON, toml string) string {
	t.Helper()
	d := acceptanceDir(t, roadmapJSON)
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	return d
}

// rdsGenerate runs `roadmap generate` in dir and returns ROADMAP.md bytes.
func rdsGenerate(t *testing.T, bin, dir string) []byte {
	t.Helper()
	out, code := runCent(t, bin, dir, "roadmap", "generate")
	if code != 0 {
		t.Fatalf("generate exit=%d\n%s", code, out)
	}
	data, err := os.ReadFile(filepath.Join(dir, "ROADMAP.md"))
	if err != nil {
		t.Fatalf("ROADMAP.md not written: %v", err)
	}
	return data
}

// rdsValidate runs `validate` and returns combined output + exit code.
func rdsValidate(t *testing.T, bin, dir string) (string, int) {
	t.Helper()
	return runCent(t, bin, dir, "validate")
}

// sampleRoadmap is a small fully-featured roadmap used across render scenarios.
const sampleRoadmap = `{"intro":"Principle line.\n\nStatus line.",
"phases":[
 {"name":"✅ Phase 0: Bootstrap","note":"Para one.\n\nPara two.",
  "features":[{"name":"setup","description":"Wire it up.","fixes":"broken hook"}]},
 {"name":"Backlog","features":[
  {"name":"f-defer","summary":"deferred bit","deferredAt":"2026-01-01",
   "source":{"feature":"feat","role":"qa"}}]}]}`

func mustHave(t *testing.T, hay, needle string) {
	t.Helper()
	if !strings.Contains(hay, needle) {
		t.Fatalf("expected %q in:\n%s", needle, hay)
	}
}
