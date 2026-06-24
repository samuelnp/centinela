package acceptance_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// cdpRepo builds a temp git repo (optionally with an origin remote) holding a
// minimal centinela.toml, and returns its path.
func cdpRepo(t *testing.T, withOrigin bool) string {
	t.Helper()
	dir := t.TempDir()
	gitIn := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	gitIn("init")
	if withOrigin {
		gitIn("remote", "add", "origin", "https://example.com/x.git")
	}
	writeFile(t, dir, "centinela.toml", "[workflow]\ndisable_auto_commit=true\nuse_worktrees=true\n")
	return dir
}

// cdpWorkflow writes a workflow state file. When atDocs is true the feature sits
// at the final step with prior steps done (so `complete` advances to done) and a
// changelog is dropped; otherwise it is a generic loadable state. worktree sets
// worktreePath so the delivery matrix sees worktree mode.
func cdpWorkflow(t *testing.T, dir, feature string, atDocs, worktree bool) {
	t.Helper()
	wp := ""
	if worktree {
		wp = ".worktrees/" + feature
	}
	step := "done"
	steps := `{}`
	if atDocs {
		step = "docs"
		steps = `{"plan":{"status":"done"},"code":{"status":"done"},"tests":{"status":"done"},"validate":{"status":"done"},"docs":{"status":"in-progress"}}`
		writeFile(t, dir, ".workflow/"+feature+"-changelog.md", "- feat: "+feature+"\n")
	}
	js := fmt.Sprintf(`{"feature":%q,"currentStep":%q,"worktreePath":%q,"profile":"strict","steps":%s}`,
		feature, step, wp, steps)
	writeFile(t, dir, filepath.Join(".workflow", feature+".json"), js)
}

// runDeliverBin runs `centinela deliver <feature> [args...]`.
func runDeliverBin(t *testing.T, dir, feature string, args ...string) (string, int) {
	t.Helper()
	return runCent(t, buildCent(t), dir, append([]string{"deliver", feature}, args...)...)
}
