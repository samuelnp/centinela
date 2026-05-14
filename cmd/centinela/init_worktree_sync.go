package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

// syncWorktreeIgnores patches ignore files for worktree-aware projects.
// Idempotent — silent when nothing changes; prints a confirmation when it does.
func syncWorktreeIgnores(repo string) error {
	res, err := worktree.SyncIgnores(repo)
	if err != nil {
		return fmt.Errorf("worktree ignore sync failed: %w", err)
	}
	for _, name := range res.Touched {
		fmt.Println(ui.RenderSuccess("worktree-ignore  " + name))
	}
	return nil
}
