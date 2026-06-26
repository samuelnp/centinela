package acceptance_test

// Acceptance: specs/lean-evidence-footprint.feature

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// leanRepo creates a temp git repo carrying the project's real .gitignore.
func leanRepo(t *testing.T) string {
	t.Helper()
	gi, err := os.ReadFile(filepath.Join(repoRoot(t), ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), gi, 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	mkWorkflow(t, dir, "demo-qa-senior.json", "demo-qa-senior.lock",
		"demo-qa-senior.md", "roadmap.json")
	return dir
}

func mkWorkflow(t *testing.T, dir string, names ...string) {
	t.Helper()
	wf := filepath.Join(dir, ".workflow")
	if err := os.MkdirAll(wf, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(wf, n), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// Scenario: Readable narratives + roadmap track; json/lock are ignored.
func TestAccEvidenceIgnoreMatrix(t *testing.T) {
	dir := leanRepo(t)
	mustGit(t, dir, "add", "-A")
	staged := mustGit(t, dir, "diff", "--cached", "--name-only")
	for _, keep := range []string{".workflow/demo-qa-senior.md", ".workflow/roadmap.json"} {
		if !strings.Contains(staged, keep) {
			t.Errorf("expected %q tracked, staged=%q", keep, staged)
		}
	}
	for _, drop := range []string{".workflow/demo-qa-senior.json", ".workflow/demo-qa-senior.lock"} {
		if strings.Contains(staged, drop) {
			t.Errorf("expected %q ignored, but it staged", drop)
		}
	}
}

// Scenario: Already-committed plumbing is untracked retroactively (local kept).
func TestAccRetroactiveUntrack(t *testing.T) {
	dir := leanRepo(t)
	mustGit(t, dir, "add", "-f", ".workflow/demo-qa-senior.json")
	mustGit(t, dir, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "seed")
	mustGit(t, dir, "rm", "--cached", "-q", ".workflow/demo-qa-senior.json")
	if tracked := mustGit(t, dir, "ls-files", ".workflow/demo-qa-senior.json"); tracked != "" {
		t.Errorf("expected untracked, ls-files=%q", tracked)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow/demo-qa-senior.json")); err != nil {
		t.Errorf("local file must survive --cached removal: %v", err)
	}
}
