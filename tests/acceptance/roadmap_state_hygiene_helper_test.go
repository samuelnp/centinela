// Acceptance: specs/roadmap-state-hygiene.feature
//
// Binary-driven end-to-end coverage for roadmap-state-hygiene (docs/plans/
// roadmap-state-hygiene.md). Every fixture is a REAL local git repo — never a
// network remote, and never the live repo's own .workflow/roadmap.json (S6's
// canonical renderer would reformat it on any mutation).
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// rshBaseRoadmap seeds three schedulable phases (one deliberately empty for
// `phase remove`), each with features to add/remove/edit/move/reorder, plus a
// Backlog finding ready to promote.
const rshBaseRoadmap = `{"phases":[
 {"name":"Phase 1","features":[
  {"name":"feature-a","description":"First thing."},
  {"name":"feature-b","description":"Second thing."},
  {"name":"feature-c","description":"Third thing."}]},
 {"name":"Phase 2","features":[]},
 {"name":"Empty Phase","features":[]},
 {"name":"Backlog","features":[
  {"name":"worth-doing","summary":"a promotable finding",
   "deferredAt":"2026-01-01T00:00:00Z","source":{"feature":"feat","role":"qa"}}]}
]}`

// rshRepo seeds a real one-commit repository with roadmap.json, so mutation
// commands run against git's actual partial-commit behavior.
func rshRepo(t *testing.T, roadmapJSON string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "rsh@centinela.dev")
	runGit(t, dir, "config", "user.name", "RSH")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), roadmapJSON)
	commit(t, dir, "seed")
	return dir
}

// rshGitOut runs git in dir and returns trimmed combined output, failing the
// test on a non-zero exit.
func rshGitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// rshChangedPaths returns the sorted, non-empty set of paths rev's commit
// touched.
func rshChangedPaths(t *testing.T, dir, rev string) []string {
	t.Helper()
	out := rshGitOut(t, dir, "show", "--name-only", "--pretty=format:", rev)
	var paths []string
	for _, ln := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(ln); s != "" {
			paths = append(paths, s)
		}
	}
	return paths
}

// rshCommitCount reports how many commits are reachable from HEAD.
func rshCommitCount(t *testing.T, dir string) int {
	t.Helper()
	n, err := strconv.Atoi(rshGitOut(t, dir, "rev-list", "--count", "HEAD"))
	if err != nil {
		t.Fatalf("rev-list --count: %v", err)
	}
	return n
}

// rshLastMsg returns HEAD's subject line.
func rshLastMsg(t *testing.T, dir string) string {
	t.Helper()
	return rshGitOut(t, dir, "log", "-1", "--pretty=%s")
}
