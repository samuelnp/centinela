package workflow

import (
	"fmt"
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

// writeReport replaces path atomically so an interrupted stamp can never leave
// a half-written verdict on disk.
func writeReport(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".stamp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
