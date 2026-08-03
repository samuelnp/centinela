package workflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/gatereport"
	"github.com/samuelnp/centinela/internal/treestate"
)

// StampVerification rewrites the verifier report's centinela:verification
// block with the current tree state, creating the block when absent and
// leaving the recorded commands untouched. It is the verifier's LAST action:
// stamping records what was verified, it never asserts that anything passed.
func StampVerification(feature, root string, run treestate.Runner) (treestate.Snapshot, error) {
	path := filepath.Join(root, GatekeeperReportPath(feature))
	data, err := os.ReadFile(path)
	if err != nil {
		return treestate.Snapshot{}, fmt.Errorf("gatekeeper report not found: %s — run `centinela artifact new %s gatekeeper` first", GatekeeperReportPath(feature), feature)
	}
	snapshot, err := treestate.Stamp(root, run)
	if err != nil {
		return treestate.Snapshot{}, err
	}
	stamped, err := gatereport.Stamped(string(data), snapshot.Revision, snapshot.Digest)
	if err != nil {
		return treestate.Snapshot{}, fmt.Errorf("%s: %w", GatekeeperReportPath(feature), err)
	}
	if err := writeReport(path, []byte(stamped)); err != nil {
		return treestate.Snapshot{}, err
	}
	return snapshot, nil
}

// reportFileMode is the mode a stamped verifier report is kept at: the same
// world-readable mode `centinela artifact new` creates it with. The hand-rolled
// replace this delegation removed used os.CreateTemp without a Chmod, so every
// `artifact stamp` silently downgraded the report to 0600.
const reportFileMode fs.FileMode = 0o644

// writeReport replaces path atomically so an interrupted stamp can never leave
// a half-written verdict on disk. It delegates to WriteFileAtomic rather than
// hand-rolling a second, weaker copy of it one file away: that copy had no
// chmod (0600 reports), no fsync and no directory fsync.
func writeReport(path string, data []byte) error {
	return WriteFileAtomic(path, data, reportFileMode)
}
