package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	analyzeBinOnce sync.Once
	analyzeBin     string
	analyzeBinErr  string
)

// buildAnalyzeBin compiles centinela once into a persistent temp dir so every
// analyze scenario shares one real binary (not a per-test t.TempDir).
func buildAnalyzeBin(t *testing.T) string {
	t.Helper()
	analyzeBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-analyze-bin")
		if err != nil {
			analyzeBinErr = err.Error()
			return
		}
		analyzeBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", analyzeBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			analyzeBinErr = err.Error() + "\n" + string(out)
			analyzeBin = ""
		}
	})
	if analyzeBin == "" {
		t.Fatalf("build centinela: %s", analyzeBinErr)
	}
	return analyzeBin
}

// runAnalyzeBin runs `centinela analyze [args...]` in dir.
func runAnalyzeBin(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildAnalyzeBin(t), dir, append([]string{"analyze"}, args...)...)
}

// analyzeGoRepo creates a minimal Go module fixture with a Makefile test target.
func analyzeGoRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	writeFile(t, dir, "go.mod", "module fixturemod\n\ngo 1.21\n")
	writeFile(t, dir, "b/b.go", "package b\n\nfunc B() {}\n")
	writeFile(t, dir, "a/a.go", "package a\n\nimport _ \"fixturemod/b\"\n")
	writeFile(t, dir, "Makefile", "test:\n\tgo test ./...\n")
	return dir
}
