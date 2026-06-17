package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	doctorBinOnce sync.Once
	doctorBin     string
	doctorBinErr  string
)

// buildDoctorBin compiles centinela once into a persistent temp dir (NOT a
// t.TempDir, which is cleaned per-test) so every scenario shares one binary.
func buildDoctorBin(t *testing.T) string {
	t.Helper()
	doctorBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-doctor-bin")
		if err != nil {
			doctorBinErr = err.Error()
			return
		}
		bin := filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			doctorBinErr = err.Error() + "\n" + string(out)
			return
		}
		doctorBin = bin
	})
	if doctorBin == "" {
		t.Fatalf("build centinela: %s", doctorBinErr)
	}
	return doctorBin
}

// runDoctor runs `doctor [args...]` in dir, returning combined output and exit
// code (0 ok, 1 on ERROR diagnoses). Reuses runCent from the shared pool.
func runDoctor(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	bin := buildDoctorBin(t)
	return runCent(t, bin, dir, append([]string{"doctor"}, args...)...)
}

// doctorRepo creates a minimal non-git project dir with a .workflow/ dir.
func doctorRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}
