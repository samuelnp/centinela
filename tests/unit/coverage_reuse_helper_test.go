package unit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// coverageScriptAbs returns the absolute path of this repo's
// scripts/check-coverage.sh so tests can invoke it from a foreign cwd.
func coverageScriptAbs(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(repoRoot(), "scripts", "check-coverage.sh"))
	if err != nil {
		t.Fatalf("abs script path: %v", err)
	}
	return p
}

// newTempGoModule scaffolds a tiny hermetic Go module (no deps, no network)
// with one covered function and one passing test.
func newTempGoModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod": "module tempmod\n\ngo 1.25\n",
		"calc.go": "package tempmod\n\n" +
			"func Add(a, b int) int { return a + b }\n",
		"calc_test.go": "package tempmod\n\nimport \"testing\"\n\n" +
			"func TestAdd(t *testing.T) {\n" +
			"\tif Add(1, 2) != 3 {\n\t\tt.Fatal(\"bad sum\")\n\t}\n}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// addFailingTest drops a test that always fails into the module.
func addFailingTest(t *testing.T, dir string) {
	t.Helper()
	src := "package tempmod\n\nimport \"testing\"\n\n" +
		"func TestAlwaysFails(t *testing.T) {\n" +
		"\tt.Fatal(\"added after the profile was recorded\")\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "fail_test.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write failing test: %v", err)
	}
}

// writeProfile runs the module's (currently passing) suite once to record
// a real coverage profile named profile.out inside dir.
func writeProfile(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "test", "./...", "-coverprofile=profile.out")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("recording profile should pass, err=%v out=%s", err, out)
	}
}
