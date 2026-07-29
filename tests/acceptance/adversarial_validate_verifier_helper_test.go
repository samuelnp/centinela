// Acceptance: specs/adversarial-validate-verifier.feature
//
// Binary-driven end-to-end coverage for the adversarial verifier gate
// (docs/plans/adversarial-validate-verifier.md Slice 7). Every fixture is a
// REAL local git repo — origin is never a network remote (a real push hangs
// go test for hours and times out claim verification).
package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var avvBinOnce sync.Once
var avvBin string

// avvBuildBin compiles the centinela binary ONCE for the whole file set and
// reuses it across every scenario, keeping this suite's wall-clock tight.
func avvBuildBin(t *testing.T) string {
	t.Helper()
	avvBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "cent-avv-bin")
		if err != nil {
			t.Fatal(err)
		}
		avvBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", avvBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return avvBin
}

// avvFixture creates a real git repo with one committed source file, so
// `git rev-parse HEAD` and a working-tree digest are both meaningful.
func avvFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "avv@centinela.dev")
	runGit(t, dir, "config", "user.name", "AVV")
	mustWrite(t, filepath.Join(dir, "src.go"), "package x\n")
	commit(t, dir, "baseline")
	return dir
}

// avvStamp runs the real `artifact stamp` subcommand so revision/treeDigest
// are computed by the production code path, never hand-derived in the test.
func avvStamp(t *testing.T, bin, dir, feature string) {
	t.Helper()
	if out, code := runCent(t, bin, dir, "artifact", "stamp", feature); code != 0 {
		t.Fatalf("artifact stamp %s failed (%d): %s", feature, code, out)
	}
}

// avvComplete runs `centinela complete <feature>` and returns combined
// output and exit code.
func avvComplete(t *testing.T, bin, dir, feature string) (string, int) {
	t.Helper()
	return runCent(t, bin, dir, "complete", feature)
}

// readFile reads path and fails the test on error.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
