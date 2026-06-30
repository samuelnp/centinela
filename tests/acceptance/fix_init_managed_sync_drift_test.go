package acceptance_test

// Acceptance: specs/fix-init-managed-sync-drift.feature

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var initDriftBinOnce sync.Once
var initDriftBin string

func buildInitDriftBin(t *testing.T) string {
	t.Helper()
	initDriftBinOnce.Do(func() {
		dir, _ := os.MkdirTemp("", "cent-initdrift-bin")
		initDriftBin = filepath.Join(dir, "centinela")
		c := exec.Command("go", "build", "-o", initDriftBin, "./cmd/centinela")
		c.Dir = repoRoot(t)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, out)
		}
	})
	return initDriftBin
}

// Scenario: a fresh init leaves no pending migration (the regression).
func TestAccInitLeavesNoPendingMigration(t *testing.T) {
	bin := buildInitDriftBin(t)
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if out, code := runCent(t, bin, dir, "init"); code != 0 {
		t.Fatalf("init exited %d: %s", code, out)
	}
	out, code := runCent(t, bin, dir, "migrate")
	if code != 0 {
		t.Fatalf("migrate exited %d: %s", code, out)
	}
	if strings.Contains(out, "update:") || strings.Contains(out, "create:") {
		t.Fatalf("fresh init should leave 0 pending migrations, got:\n%s", out)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.HasPrefix(string(b), "<!-- centinela:managed-version=") {
		t.Errorf("AGENTS.md missing managed-version header after init")
	}
}
