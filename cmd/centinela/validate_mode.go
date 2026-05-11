package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// gitDiffResolver is overridable by tests.
var gitDiffResolver = gitdiff.Default

// currentEnv reads the environment signals that affect mode resolution.
// CI is detected via the universal `CI` env var (set by every major
// CI system: GitHub Actions, GitLab, CircleCI, Travis, etc.).
func currentEnv() config.Env {
	v := os.Getenv("CI")
	return config.Env{CI: v == "true" || v == "1"}
}

// resolveDiffFilter returns the gate filter and a human-readable header
// fragment describing which mode and base are active for this run.
// On any git failure in Changed mode it degrades to Full and prints a notice.
func resolveDiffFilter(cfg *config.Config, mode config.Mode) (*gitdiff.Set, string) {
	if mode == config.ModeFull {
		return nil, "(full scan)"
	}
	set, summary, _ := gitDiffResolver.ChangedFiles(cfg.Validate.DiffBase, true)
	if summary.Degrade != "" {
		fmt.Println("notice: diff-aware degraded to full scan — " + summary.Degrade)
		return nil, "(full scan)"
	}
	return set, fmt.Sprintf("(diff-aware: %d files changed since %s)", summary.Files, summary.Base)
}
