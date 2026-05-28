package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildCent compiles the centinela binary once per test and returns its path.
func buildCent(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "cent")
	c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	c.Dir = repoRoot(t)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("build cent: %v\n%s", err, out)
	}
	return bin
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd() // tests/acceptance
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

// runCent runs the binary in dir and returns combined output + exit code.
func runCent(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return string(out), code
}

func acceptanceDir(t *testing.T, roadmapJSON string) string {
	t.Helper()
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if roadmapJSON != "" {
		if err := os.WriteFile(filepath.Join(d, ".workflow", "roadmap.json"), []byte(roadmapJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func seedDoneAt(t *testing.T, dir, feature string) {
	t.Helper()
	body := `{"feature":"` + feature + `","currentStep":"done","steps":{}}`
	if err := os.WriteFile(filepath.Join(dir, ".workflow", feature+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
