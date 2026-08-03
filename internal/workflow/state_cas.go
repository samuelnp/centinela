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
// Residual TOCTOU: two processes can both pass this check and both rename,
// microseconds apart. Closing that needs a short flock held across re-read →
// rename; the detection above is what turns the observed silent data loss into
// a loud, actionable failure.
func checkNotStale(path string, wf *Workflow, current []byte) error {
	if wf.loadedDigest == "" || wf.loadedDigest == digestOf(current) {
		return nil
	}
	return fmt.Errorf("%s changed on disk since this command read it — another "+
		"centinela process wrote it (a concurrent `route set`, `complete`, or "+
		"`revise`). Refusing to write so that update is not lost. Re-run this "+
		"command to apply your change on top of the current state", path)
}
