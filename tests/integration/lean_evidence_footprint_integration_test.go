package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// seedIgnoreRepo creates a temp git repo whose .gitignore is the repo's own,
// so the test exercises the shipped patterns end to end.
func seedIgnoreRepo(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	gi, err := os.ReadFile(filepath.Join(wd, "..", "..", ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), gi, 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatalf("mkdir .workflow: %v", err)
	}
	return dir
}

// ignored reports whether git ignores rel within dir (exit 0 = ignored).
func ignored(t *testing.T, dir, rel string) bool {
	t.Helper()
	return exec.Command("git", "-C", dir, "check-ignore", "-q", rel).Run() == nil
}

func TestEvidencePlumbingIgnored(t *testing.T) {
	dir := seedIgnoreRepo(t)
	for _, rel := range []string{
		".workflow/f-big-thinker.json",
		".workflow/f-senior-engineer.json",
		".workflow/f-big-thinker.lock",
	} {
		if !ignored(t, dir, rel) {
			t.Errorf("expected %q to be ignored", rel)
		}
	}
}

func TestKbAndRoadmapNotIgnored(t *testing.T) {
	dir := seedIgnoreRepo(t)
	// Durable state must stay tracked: a bare per-feature ledger (.workflow/
	// f.json) is NOT role-suffixed, so the fail-safe policy keeps it tracked.
	for _, rel := range []string{
		".workflow/roadmap.json",
		".workflow/roadmap-analysis.json",
		".workflow/audit-baseline.json",
		".workflow/f.json",
		".workflow/f-big-thinker.md",
	} {
		if ignored(t, dir, rel) {
			t.Errorf("expected %q to remain tracked (not ignored)", rel)
		}
	}
}
