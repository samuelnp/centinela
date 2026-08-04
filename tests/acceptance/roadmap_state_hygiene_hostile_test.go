// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os"
	"strings"
	"testing"
)

// Scenario: a hostile git environment warns but never fails the mutation
func TestRsh_HostileGitEnvironmentWarnsNeverFails(t *testing.T) {
	bin := buildCent(t)
	cases := []struct {
		name, reason string
		hostile      func(t *testing.T, dir string)
	}{
		{"not a git repository", "no git repository", func(t *testing.T, dir string) {
			mustRemoveAll(t, dir+"/.git")
		}},
		{"no commits yet (no HEAD)", "no HEAD", func(t *testing.T, dir string) {
			mustRemoveAll(t, dir+"/.git")
			runGit(t, dir, "init", "-q", "-b", "main")
		}},
		{"a merge in progress", "merge in progress", func(t *testing.T, dir string) {
			head := rshGitOut(t, dir, "rev-parse", "HEAD")
			mustWrite(t, dir+"/.git/MERGE_HEAD", head+"\n")
		}},
		{"a rebase in progress", "rebase in progress", func(t *testing.T, dir string) {
			if err := os.MkdirAll(dir+"/.git/rebase-apply", 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{"git commit fails", "commit failed", func(t *testing.T, dir string) {
			runGit(t, dir, "config", "commit.gpgsign", "true")
			runGit(t, dir, "config", "gpg.program", "/nonexistent-gpg-binary")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := rshRepo(t, rshBaseRoadmap)
			tc.hostile(t, dir)

			out, code := runCent(t, bin, dir, "roadmap", "defer", "resilient-thing", "--summary", "x")
			if code != 0 {
				t.Fatalf("mutation must still exit 0: %d\n%s", code, out)
			}
			body := mustRead(t, dir+"/.workflow/roadmap.json")
			if !strings.Contains(body, "resilient-thing") {
				t.Fatalf("the finding must be on disk: %s", body)
			}
			if _, err := os.Stat(dir + "/ROADMAP.md"); err != nil {
				t.Fatalf("ROADMAP.md must be regenerated: %v", err)
			}
			// AC5: the warning must name the reason and say the state was not
			// committed. It may only add "in your working tree" because the
			// read-back verified the record really is there (F1).
			containsAll(t, out, "not committed", "in your working tree", tc.reason)
		})
	}
}

func mustRemoveAll(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
