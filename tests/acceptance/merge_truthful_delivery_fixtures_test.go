package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mtdfBarePrimaryRepo clones an ordinary fixture into a LOCAL bare repository
// and registers a linked worktree against it, so the first entry of
// `git worktree list --porcelain` carries the `bare` attribute. Returns the
// worktree path the operator would be standing in. No network: the clone
// source is a local directory.
func mtdfBarePrimaryRepo(t *testing.T, feature string) string {
	t.Helper()
	src := mergeRepo(t, feature, false)
	work := t.TempDir()
	bare := filepath.Join(work, "origin.git")
	mtdaGit(t, work, "clone", "--bare", "-q", src, bare)
	wt := filepath.Join(work, feature)
	mtdaGit(t, bare, "worktree", "add", "-q", wt, feature)
	return wt
}

// mtdfValidateFailRepo builds a repo whose merged result genuinely fails
// `centinela validate`: main carries a validate command that exits non-zero,
// and the feature branch merges cleanly on top of it.
func mtdfValidateFailRepo(t *testing.T, feature string) string {
	t.Helper()
	d := t.TempDir()
	for _, a := range [][]string{{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@centinela.dev"}, {"config", "user.name", "QA"}} {
		mtdaGit(t, d, a...)
	}
	writeFile(t, d, ".gitignore", ".worktrees/\n.workflow/\n")
	writeFile(t, d, "centinela.toml",
		"[validate]\ncommands = [\"exit 1\"]\n[gates]\nfile_size = false\n")
	writeFile(t, d, "shared.txt", "base\n")
	mtdaGit(t, d, "add", ".")
	mtdaGit(t, d, "commit", "-q", "-m", "base")
	wt := filepath.Join(d, ".worktrees", feature)
	mtdaGit(t, d, "worktree", "add", "-q", filepath.Join(".worktrees", feature), "-b", feature)
	writeFile(t, wt, "feature.txt", feature+"\n") // no conflict with main
	mtdaGit(t, wt, "add", ".")
	mtdaGit(t, wt, "commit", "-q", "-m", "feature edit")
	return d
}

// mtdfNoOpMergeGit writes a `git` shim that reports success for
// `git merge --no-ff <branch>` WITHOUT moving HEAD, delegating every other
// subcommand to the real binary. It manufactures the one shape real git makes
// unreachable: a merge that exits 0 while the ref did not advance and the
// branch is not an ancestor. Returns a PATH-prefixed environment.
func mtdfNoOpMergeGit(t *testing.T) []string {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	dir := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = \"merge\" ] && [ \"$2\" = \"--no-ff\" ]; then\n" +
		"  echo 'Already up to date.'\n  exit 0\nfi\nexec " + realGit + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// runCentEnv is runCent with an explicit environment.
func runCentEnv(t *testing.T, bin, dir string, env []string, args ...string) (string, int) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Dir = dir
	c.Env = env
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return string(out), code
}
