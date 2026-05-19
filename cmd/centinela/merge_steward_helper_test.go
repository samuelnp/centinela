package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// stewardRepo builds a hermetic git repo with a committed feature worktree.
// chdir is the caller's responsibility (restore via the returned cleanup).
func stewardRepo(t *testing.T, feature string, conflict bool) string {
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

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
}

// writeStewardEvidence writes a schema-valid merge-steward evidence pair
// (.md + .json) with the given handoffTo so stewardEvidenceValidator passes.
func writeStewardEvidence(t *testing.T, feature, handoffTo string) {
	t.Helper()
	_ = os.MkdirAll(".workflow", 0o755)
	md := ".workflow/" + feature + "-merge-steward.md"
	_ = os.WriteFile(md, []byte("# steward report\nresolved\n"), 0o644)
	ts := time.Now().UTC().Format(time.RFC3339)
	js := `{"feature":"` + feature + `","step":"merge","role":"merge-steward",` +
		`"status":"done","generatedAt":"` + ts + `",` +
		`"inputs":[".workflow/` + feature + `-merge-pending.json"],` +
		`"outputs":["` + md + `"],"edgeCases":["text-conflict"],` +
		`"handoffTo":"` + handoffTo + `"}`
	_ = os.WriteFile(".workflow/"+feature+"-merge-steward.json", []byte(js), 0o644)
}
