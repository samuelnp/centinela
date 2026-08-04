package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// syncGitRepo chdirs into a real one-commit repository seeded with a roadmap,
// so the assertions run against git's actual partial-commit behavior.
func syncGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"Phase 1","features":[{"name":"a","dependsOn":[]}]}]}`
	if err := os.WriteFile(".workflow/roadmap.json", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"}, {"config", "user.email", "t@e.com"},
		{"config", "user.name", "T"}, {"config", "commit.gpgsign", "false"},
		{"add", "-A"}, {"commit", "-q", "-m", "seed"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func git(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

// AC1+AC2+AC3 end to end through the real command.
func TestSyncRoadmapStateCommitsOnlyItsPathspec(t *testing.T) {
	syncGitRepo(t)
	if err := os.WriteFile("other.txt", []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, "add", "other.txt")
	deferSummary, deferSource = "a finding", "feat/eng"
	if err := runRoadmapDefer(nil, []string{"flaky-thing"}); err != nil {
		t.Fatal(err)
	}
	changed := strings.Fields(git(t, "show", "--name-only", "--pretty=format:", "HEAD"))
	if strings.Join(changed, ",") != ".workflow/roadmap.json,ROADMAP.md" {
		t.Fatalf("commit changed %v", changed)
	}
	if got := strings.TrimSpace(git(t, "log", "-1", "--pretty=%s")); got != "chore(roadmap): defer flaky-thing" {
		t.Fatalf("message = %q", got)
	}
	if got := strings.TrimSpace(git(t, "diff", "--cached", "--name-only")); got != "other.txt" {
		t.Fatalf("unrelated staged work = %q, want it still staged", got)
	}
	if strings.TrimSpace(git(t, "status", "--porcelain", "--", ".workflow/roadmap.json", "ROADMAP.md")) != "" {
		t.Fatal("roadmap state must be clean after the mutation")
	}
}

// AC4: the knob skips the COMMIT, never the REGENERATION.
func TestSyncRoadmapStateHonorsDisableAutoCommit(t *testing.T) {
	syncGitRepo(t)
	if err := os.WriteFile("centinela.toml", []byte("[workflow]\ndisable_auto_commit = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if roadmapAutoCommitEnabled() {
		t.Fatal("disable_auto_commit must switch auto-commit off")
	}
	before := strings.TrimSpace(git(t, "rev-parse", "HEAD"))
	deferSummary, deferSource = "x", "feat/eng"
	if err := runRoadmapDefer(nil, []string{"policy-thing"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(git(t, "rev-parse", "HEAD")) != before {
		t.Fatal("no commit may be created when auto-commit is disabled")
	}
	if _, err := os.Stat("ROADMAP.md"); err != nil {
		t.Fatalf("ROADMAP.md must still be regenerated: %v", err)
	}
}
