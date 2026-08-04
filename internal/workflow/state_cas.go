package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// digestOf fingerprints the RAW file bytes rather than the parsed struct, so
// any change another writer made counts as a conflict — including changes to
// fields this binary does not model.
func digestOf(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// checkNotStale refuses a save whose in-memory workflow was read from a version
// of the file that no longer exists on disk.
//
// A workflow that was never loaded — built by New/NewWithOrder, as `centinela
// start` and the autostart hook do — carries no digest and is never checked.
// There is nothing it could be clobbering, and that exemption is what keeps
// every existing hand-authored test fixture working untouched.
//
// This is optimistic concurrency on purpose, and the alternatives are worse:
//
//   - A LOCK cannot help. The stale read happens minutes before the write —
//     `complete` loads the workflow, runs the whole validate gate, then saves —
//     so serializing writers would block a concurrent `route set` for the
//     length of a gate run, far past evidence.LockTimeout. Narrowing the lock
//     to the write itself leaves the lost update entirely intact, because the
//     data was already stale when the lock was taken.
//   - An IN-PROCESS MUTEX guards nothing. `route set` and `complete` are
//     separate OS processes.
//
// What this DOES and DOES NOT guarantee — measured, not estimated:
//
//   - It detects STALENESS, not concurrency. A writer whose read is older than
//     the file on disk is refused. That is the minutes-wide read-modify-write
//     window this feature exists to close (`complete` loads, runs a whole
//     validate gate, then saves), and there detection is perfect: with saves
//     serialised, 24 competing writers produced 1 success, 23 refusals and 0
//     lost updates.
//   - It does NOT protect two saves that genuinely OVERLAP. The window between
//     this check and the rename is the marshal + CreateTemp + write + fsync +
//     rename — MILLISECONDS, not microseconds, and empirically wide enough to
//     matter: 24 fully-overlapping writers produced 24 reported successes, 0
//     refusals and 23 silently lost updates. Closing that needs a short flock
//     held across re-read → rename (deferred: close-state-save-toctou).
//
// So: a stale read is a loud, actionable failure; a true write-write race is
// still last-writer-wins. Do not read the refusal message as a promise of
// mutual exclusion.
func checkNotStale(path string, wf *Workflow, current []byte) error {
	if wf.loadedDigest == "" || wf.loadedDigest == digestOf(current) {
		return nil
	}
	return fmt.Errorf("%s changed on disk since this command read it — another "+
		"centinela process wrote it (a concurrent `route set`, `complete`, or "+
		"`revise`). Refusing to write so that update is not lost. Re-run this "+
		"command to apply your change on top of the current state", path)
}

// checkNotDeleted refuses a save whose target vanished between Load and Save.
//
// A missing file is a first write ONLY for a workflow nobody read (New /
// NewWithOrder, i.e. `start` and the autostart hook). For a LOADED workflow it
// means another process removed the state file — an abandoned feature, a
// `git clean -xfd`, a worktree teardown, a `git checkout` landing during a
// minutes-long `complete`. Recreating it there silently resurrects a workflow
// someone deliberately deleted, with content that is by definition stale, so it
// is reported exactly like any other conflict: refuse, and let the operator
// re-run.
func checkNotDeleted(path string, wf *Workflow) error {
	if wf.loadedDigest == "" {
		return nil
	}
	return fmt.Errorf("%s was deleted since this command read it — another "+
		"process removed the state file (an abandoned workflow, a worktree "+
		"cleanup, a `git checkout`). Refusing to write so a deleted workflow is "+
		"not silently resurrected. Re-run `centinela start %s` if you meant to "+
		"begin again", path, wf.Feature)
}
