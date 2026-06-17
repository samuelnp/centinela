package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samuelnp/centinela/internal/worktree"
)

// gitRunner is overridable for tests; default shells out to git in repo.
var gitRunner = func(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	return cmd.CombinedOutput()
}

// gitAvailable reports whether repo is inside a git work tree.
func gitAvailable(repo string) bool {
	out, err := gitRunner(repo, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// listWorktrees returns the sorted feature directory names under .worktrees/.
func listWorktrees(repo string) []string {
	entries, err := os.ReadDir(filepath.Join(repo, worktree.Dir))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// branchMerged reports whether the feature's branch is merged into main. A
// missing branch counts as merged (its work is gone). Unknown main → false.
func branchMerged(repo, feature string) bool {
	if _, err := gitRunner(repo, "rev-parse", "--verify", "refs/heads/"+feature); err != nil {
		return true
	}
	out, err := gitRunner(repo, "branch", "--merged", "main")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "* ")) == feature {
			return true
		}
	}
	return false
}

// workflowDone reports whether the feature's workflow JSON exists at the repo
// root AND is at the terminal "done" step. A missing file is NOT treated as
// done: under the worktree flow a feature's state lives on its branch and may
// be absent from the canonical checkout while the work is still in progress.
func workflowDone(repo, feature string) bool {
	data, err := os.ReadFile(filepath.Join(repo, ".workflow", feature+".json"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), `"currentStep": "done"`)
}
