// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Scenario: concurrent mutations entering by different path spellings all survive
//
// The probe combination the unit tests miss: the in-process symlink test never
// exercises git's index.lock, and the multi-process test never uses a symlink.
// This is both at once — real `centinela roadmap defer` PROCESSES, half of them
// entering the repo through a symlinked directory. That combination is what
// destroyed 10 of 24 records when the lock was keyed on an unresolved path.
func TestRsh_ConcurrentDefersAcrossASymlinkedPath(t *testing.T) {
	bin := buildCent(t)
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	seedRepoAt(t, real)
	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	slugs := map[string]string{
		"via-real-1": real, "via-real-2": real,
		"via-link-1": link, "via-link-2": link,
	}
	var wg sync.WaitGroup
	for slug, dir := range slugs {
		wg.Add(1)
		go func(slug, dir string) {
			defer wg.Done()
			c := exec.Command(bin, "roadmap", "defer", slug, "--summary", "concurrent")
			c.Dir = dir
			if out, err := c.CombinedOutput(); err != nil {
				t.Errorf("worker %s: %v\n%s", slug, err, out)
			}
		}(slug, dir)
	}
	wg.Wait()

	body := mustRead(t, filepath.Join(real, ".workflow", "roadmap.json"))
	for slug := range slugs {
		if !strings.Contains(body, `"`+slug+`"`) {
			t.Fatalf("%q was destroyed — a path spelling re-opened the write race:\n%s", slug, body)
		}
	}
	// Exactly ONE lock, not one per spelling.
	locks, err := filepath.Glob(filepath.Join(real, ".git", "centinela-roadmap-*.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if len(locks) != 1 {
		t.Fatalf("want exactly 1 lock file, got %d: %v", len(locks), locks)
	}
	if status := rshGitOut(t, real, "status", "--porcelain"); status != "" {
		t.Fatalf("the tree must be clean afterwards, got:\n%s", status)
	}
}

// seedRepoAt initializes a one-commit repo with a roadmap at an EXISTING dir.
func seedRepoAt(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.email", "rsh@centinela.dev")
	runGit(t, dir, "config", "user.name", "RSH")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), rshBaseRoadmap)
	commit(t, dir, "seed")
}
