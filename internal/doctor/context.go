package doctor

import (
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/worktree"
)

// NewContext resolves the canonical repo root from cwd, changes the process
// working directory to it (so the reused CWD-relative domain APIs — config.Load,
// roadmap.Load, setup, evidence.Repair — operate on the canonical repo rather
// than a worktree subtree), and loads centinela.toml.
//
// When cwd is inside .worktrees/<feature>, the repo root is the parent of the
// .worktrees directory. A centinela.toml parse error is captured (not fatal) so
// the config check can degrade to ERROR while the other checks still run.
func NewContext(cwd string) (Context, error) {
	root := resolveRoot(cwd)
	if err := os.Chdir(root); err != nil {
		return Context{}, err
	}
	cfg, err := config.Load()
	ctx := Context{Root: root}
	if err != nil {
		ctx.CfgErr = err
		return ctx, nil
	}
	ctx.Config = cfg
	return ctx, nil
}

// resolveRoot returns the canonical repo root for cwd. If cwd is inside a
// worktree, it climbs out to the parent of .worktrees; otherwise cwd itself.
func resolveRoot(cwd string) string {
	if _, wtRoot := worktree.DetectFeatureFromCwd(cwd); wtRoot != "" {
		// wtRoot == <repo>/.worktrees/<feature>; repo is two levels up.
		return filepath.Dir(filepath.Dir(wtRoot))
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return cwd
	}
	return abs
}
