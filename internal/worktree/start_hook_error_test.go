package worktree

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// MaybeProvision must propagate Create's slug error when the flag is on and
// the repo is a git repo (so the slug check is actually reached).
func TestMaybeProvision_FlagOnInvalidSlug_PropagatesError(t *testing.T) {
	repo := t.TempDir()
	if _, err := gitRunner(repo, "init", "-q", "-b", "main"); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = true
	if _, err := MaybeProvision(repo, "Bad Slug/../x", cfg); err == nil {
		t.Fatal("MaybeProvision must surface the invalid-slug error from Create")
	}
}
