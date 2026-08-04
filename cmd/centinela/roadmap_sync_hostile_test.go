package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitutil"
	"github.com/samuelnp/centinela/internal/roadmap"
)

// AC5: every hostile git environment must still leave the mutation applied and
// the command exiting 0. The reason is what changes, never the exit code.
func TestMutationSurvivesHostileGitEnvironments(t *testing.T) {
	cases := []struct {
		name, reason string
		hostile      func(t *testing.T)
	}{
		{"not a git repository", "no git repository", func(t *testing.T) {
			if err := os.RemoveAll(".git"); err != nil {
				t.Fatal(err)
			}
		}},
		{"no commits yet", "no HEAD", func(t *testing.T) {
			if err := os.RemoveAll(".git"); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("git", "init", "-q", "-b", "main").CombinedOutput(); err != nil {
				t.Fatalf("git init: %v %s", err, out)
			}
		}},
		{"merge in progress", "merge in progress", func(t *testing.T) {
			gitDir := strings.TrimSpace(git(t, "rev-parse", "--absolute-git-dir"))
			head := strings.TrimSpace(git(t, "rev-parse", "HEAD"))
			if err := os.WriteFile(gitDir+"/MERGE_HEAD", []byte(head+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"rebase in progress", "rebase in progress", func(t *testing.T) {
			gitDir := strings.TrimSpace(git(t, "rev-parse", "--absolute-git-dir"))
			if err := os.MkdirAll(gitDir+"/rebase-apply", 0o755); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			syncGitRepo(t)
			tc.hostile(t)
			deferSummary, deferSource = "x", "feat/eng"
			if err := runRoadmapDefer(nil, []string{"resilient-thing"}); err != nil {
				t.Fatalf("the mutation must still succeed: %v", err)
			}
			body, err := os.ReadFile(".workflow/roadmap.json")
			if err != nil || !strings.Contains(string(body), "resilient-thing") {
				t.Fatalf("the finding must be on disk: %v", err)
			}
			if _, err := os.Stat("ROADMAP.md"); err != nil {
				t.Fatalf("ROADMAP.md must be regenerated: %v", err)
			}
			if got := gitutil.CommitBlockedReason("."); got != tc.reason {
				t.Fatalf("reason = %q, want %q", got, tc.reason)
			}
		})
	}
}

// The adapter must translate git's sentinel into the domain's, so Sync can tell
// "nothing changed" (silent) apart from "the commit failed" (a warning).
func TestRoadmapCommitterTranslatesNothingToCommit(t *testing.T) {
	syncGitRepo(t)
	err := roadmapCommitter{repo: "."}.Commit("m", []string{"absent.json"})
	if !errors.Is(err, roadmap.ErrNoChange) {
		t.Fatalf("want roadmap.ErrNoChange, got %v", err)
	}
}

func TestRoadmapCommitterSurfacesRealFailures(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("ROADMAP.md", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := roadmapCommitter{repo: "."}.Commit("m", []string{"ROADMAP.md"})
	if err == nil || errors.Is(err, roadmap.ErrNoChange) {
		t.Fatalf("a non-repo must surface as a real failure, got %v", err)
	}
}
