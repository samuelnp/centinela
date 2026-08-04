// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import "testing"

// Scenario: a mutation inside a feature worktree commits to that worktree's branch
func TestRsh_MutationInWorktreeNeverTouchesPrimaryCheckout(t *testing.T) {
	bin := buildCent(t)
	primary := rshRepo(t, rshBaseRoadmap)
	rdsGenerate(t, bin, primary)
	commit(t, primary, "generate ROADMAP.md")

	wt := primary + "-wt"
	runGit(t, primary, "worktree", "add", "-b", "some-feature", wt, "main")

	primaryHead := rshGitOut(t, primary, "rev-parse", "HEAD")
	primaryStatus := rshGitOut(t, primary, "status", "--porcelain")

	out, code := runCent(t, bin, wt, "roadmap", "defer", "worktree-thing", "--summary", "x")
	if code != 0 {
		t.Fatalf("defer exit=%d\n%s", code, out)
	}

	if got := rshGitOut(t, wt, "branch", "--show-current"); got != "some-feature" {
		t.Fatalf("new commit must land on the worktree's branch, got %q", got)
	}
	if got := rshGitOut(t, wt, "rev-parse", "HEAD"); got == primaryHead {
		t.Fatal("the worktree's HEAD must have advanced")
	}

	if got := rshGitOut(t, primary, "rev-parse", "HEAD"); got != primaryHead {
		t.Fatalf("primary checkout's HEAD must be unchanged: was %q, now %q", primaryHead, got)
	}
	if got := rshGitOut(t, primary, "status", "--porcelain"); got != primaryStatus {
		t.Fatalf("primary checkout's working tree/index must be unchanged: was %q, now %q", primaryStatus, got)
	}
}
