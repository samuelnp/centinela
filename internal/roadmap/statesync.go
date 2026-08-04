package roadmap

import (
	"errors"

	"github.com/samuelnp/centinela/internal/roadmapstate"
)

// ErrNoChange is the sentinel a Committer returns when none of the pathspec's
// files differ from HEAD, so no commit was needed. It is a normal outcome: a
// mutation that leaves roadmap state byte-identical must create no commit.
var ErrNoChange = errors.New("roadmap state is unchanged")

// Committer commits an exact pathspec and nothing else. The domain must never
// run git itself, so the CLI supplies the implementation; this interface is the
// entire contract between them.
type Committer interface {
	// Commit creates one commit containing exactly paths and nothing else.
	// Returning ErrNoChange means the paths were already up to date.
	Commit(msg string, paths []string) error
}

// SyncOptions describes one post-mutation sync. ExtraPaths are the files this
// particular mutation ALSO rewrote (promote's analysis/quality artifacts); the
// always-present pair comes from roadmapstate, never from the caller.
type SyncOptions struct {
	Verb       string
	Subject    string
	ExtraPaths []string
	Commit     bool
	C          Committer
}

// SyncReport is what happened, for the presentation layer to render. Warn
// marks the hostile-environment cases an operator should see on stderr, as
// opposed to the ordinary "policy says don't commit" and "nothing changed".
type SyncReport struct {
	Message     string
	Paths       []string
	Regenerated bool
	Committed   bool
	Warn        bool
	Reason      string
	// InWorkingTree is a VERIFIED read-back, not an assumption: true only when
	// roadmap.json and ROADMAP.md on disk still agree after this sync. Without
	// it the "left uncommitted" line would assert the record is in your tree
	// even when a concurrent write had already replaced it.
	InWorkingTree bool
}

// Sync is THE post-mutation choke point: it regenerates ROADMAP.md from the
// just-written roadmap.json, then commits the roadmap-state pathspec.
//
// Regeneration is correctness (the roadmap_drift gate) and always runs; the
// commit is VCS policy and is skipped when opts.Commit is false. Sync never
// returns an error — every failure becomes a Reason, because a non-zero exit
// after a successful on-disk mutation would invite a retry and a duplicate
// record.
func Sync(opts SyncOptions) SyncReport {
	rep := SyncReport{
		Message: roadmapstate.Message(opts.Verb, opts.Subject),
		Paths:   append(roadmapstate.Paths(), opts.ExtraPaths...),
	}
	changed, err := RegenerateMarkdown()
	if err != nil {
		rep.Warn, rep.Reason = true, "ROADMAP.md regeneration failed: "+err.Error()
		return rep
	}
	rep.Regenerated = changed
	rep.InWorkingTree = StateInSync()
	if !opts.Commit {
		rep.Reason = "auto-commit is disabled"
		return rep
	}
	if opts.C == nil {
		rep.Warn, rep.Reason = true, "no committer configured"
		return rep
	}
	switch err := opts.C.Commit(rep.Message, rep.Paths); {
	case err == nil:
		rep.Committed = true
	case errors.Is(err, ErrNoChange):
		rep.Reason = ErrNoChange.Error()
	default:
		rep.Warn, rep.Reason = true, err.Error()
	}
	return rep
}
