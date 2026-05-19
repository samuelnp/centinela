package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// buildCentinela compiles the CLI once per test into a temp dir.
func buildCentinela(t *testing.T, work string) string {
	t.Helper()
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(work, "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela: %v\n%s", err, out)
	}
	return bin
}

// mergeRepo builds a hermetic git repo with a committed feature worktree.
// When conflict=true, main diverges so `git merge` produces a text conflict.
func mergeRepo(t *testing.T, feature string, conflict bool) string {
	t.Helper()
	d := t.TempDir()
	git := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	for _, a := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@centinela.dev"},
		{"config", "user.name", "QA"},
	} {
		git(d, a...)
	}
	_ = os.WriteFile(filepath.Join(d, "shared.txt"), []byte("base\n"), 0o644)
	_ = os.WriteFile(filepath.Join(d, ".gitignore"),
		[]byte(".worktrees/\n.workflow/\n"), 0o644)
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0o644)
	git(d, "add", ".")
	git(d, "commit", "-q", "-m", "base")
	wt := filepath.Join(d, ".worktrees", feature)
	git(d, "worktree", "add", filepath.Join(".worktrees", feature), "-b", feature)
	_ = os.WriteFile(filepath.Join(wt, "shared.txt"), []byte("feature\n"), 0o644)
	git(wt, "add", ".")
	git(wt, "commit", "-q", "-m", "feature edit")
	if conflict {
		_ = os.WriteFile(filepath.Join(d, "shared.txt"), []byte("main\n"), 0o644)
		git(d, "add", ".")
		git(d, "commit", "-q", "-m", "main edit")
	}
	return d
}

func writeMergeEvidence(t *testing.T, repo, feature, handoffTo string) {
	t.Helper()
	wf := filepath.Join(repo, ".workflow")
	_ = os.MkdirAll(wf, 0o755)
	mdRel := ".workflow/" + feature + "-merge-steward.md"
	_ = os.WriteFile(filepath.Join(wf, feature+"-merge-steward.md"),
		[]byte("# steward report\nresolved\n"), 0o644)
	ts := time.Now().UTC().Format(time.RFC3339)
	js := `{"feature":"` + feature + `","step":"merge","role":"merge-steward",` +
		`"status":"done","generatedAt":"` + ts + `",` +
		`"inputs":[".workflow/` + feature + `-merge-pending.json"],` +
		`"outputs":["` + mdRel + `"],"edgeCases":["text-conflict"],` +
		`"handoffTo":"` + handoffTo + `"}`
	_ = os.WriteFile(filepath.Join(wf, feature+"-merge-steward.json"), []byte(js), 0o644)
}

func runBin(t *testing.T, bin, dir string, args ...string) (string, error) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}

// abortRepoMerge clears the in-progress conflicted merge so the main tree
// is clean again — emulating the operator/steward applying a resolution
// before `centinela merge --continue`.
func abortRepoMerge(t *testing.T, repo string) {
	t.Helper()
	c := exec.Command("git", "merge", "--abort")
	c.Dir = repo
	_, _ = c.CombinedOutput() // best-effort: no-op if not mid-merge
}
