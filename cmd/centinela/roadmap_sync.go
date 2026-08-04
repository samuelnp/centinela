package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitutil"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

// roadmapCommitter adapts the gitutil pathspec commit to roadmap.Committer,
// translating git's "nothing to commit" into the domain's sentinel so the
// domain stays free of any git vocabulary.
type roadmapCommitter struct{ repo string }

// Commit creates one pathspec-scoped commit in the CURRENT tree. Inside
// .worktrees/<feature> that is the feature branch — deliberately: the deferral
// travels with the branch. We never reach into the primary checkout.
func (c roadmapCommitter) Commit(msg string, paths []string) error {
	err := gitutil.CommitPaths(c.repo, msg, paths)
	if errors.Is(err, gitutil.ErrNothingToCommit) {
		return roadmap.ErrNoChange
	}
	return err
}

// syncRoadmapState regenerates ROADMAP.md from the just-mutated roadmap.json
// and commits the roadmap-state pathspec. It returns nothing on purpose: the
// commit is best-effort, so a hostile git environment warns and the mutation
// still exits 0 rather than inviting a retry that would double-record.
func syncRoadmapState(verb, subject string, extra ...string) {
	rep := roadmap.Sync(roadmap.SyncOptions{
		Verb: verb, Subject: subject, ExtraPaths: extra,
		Commit: roadmapAutoCommitEnabled(), C: roadmapCommitter{repo: "."},
	})
	line := ui.RenderRoadmapSync(rep)
	// A failed read-back is louder than a warning: the record may be gone.
	if rep.Warn || !rep.InWorkingTree {
		fmt.Fprintln(os.Stderr, line)
		return
	}
	fmt.Println(line)
}

// roadmapAutoCommitEnabled reports whether workflow.disable_auto_commit allows
// the mutation commit. An unreadable config falls back to the documented
// default (auto-commit on); it must never block a mutation.
func roadmapAutoCommitEnabled() bool {
	cfg, err := config.Load()
	if err != nil {
		return true
	}
	return !cfg.Workflow.DisableAutoCommit
}
