package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
)

// runBrownBin runs `centinela roadmap brownfield [args...]` in dir against the
// shared real binary, returning combined output and the exit code.
func runBrownBin(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildAnalyzeBin(t), dir, append([]string{"roadmap", "brownfield"}, args...)...)
}

// brownDir creates a temp dir with .workflow/analysis.json seeded from body and
// returns the dir. When body is empty no inventory is written (missing-input case).
func brownDir(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if body == "" {
		return dir
	}
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".workflow", "analysis.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// goBrownInventory promotes three behavioral targets (TODO-bearing each), so the
// draft carries a Baseline phase plus a gap phase.
const goBrownInventory = `{"schemaVersion":1,"primaryLanguage":"Go",
"packages":["cmd/app","internal/handler","internal/service"],
"graph":{"kind":"go-packages","edges":[]}}`

// docOnlyBrownInventory has no behavioral packages, so zero targets, no gaps.
const docOnlyBrownInventory = `{"schemaVersion":1,"primaryLanguage":"Markdown",
"packages":["docs","readme"],"graph":{"kind":"none","edges":[]}}`
