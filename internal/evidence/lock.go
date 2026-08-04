package evidence

import (
	"fmt"
	"path/filepath"

	"github.com/samuelnp/centinela/internal/filelock"
	"github.com/samuelnp/centinela/internal/workflow"
)

// LockTimeout is the upper bound on how long Lock waits before giving up.
// Short by design — the typical CLI invocation holds the lock for a few
// hundred microseconds, so even one second is generous.
const LockTimeout = filelock.DefaultTimeout

// LockPollInterval is the retry cadence inside the LockTimeout window.
const LockPollInterval = filelock.DefaultPollInterval

// lockPath returns the .lock sibling file we lock for advisory mutual
// exclusion. Keeping it separate from the JSON means the JSON itself is
// never opened with a held lock.
func lockPath(feature string, role Role) string {
	return filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.lock", feature, role))
}

// Lock takes an advisory file lock on the (feature, role) pair and returns
// a release function that callers MUST defer. On timeout the error message
// names the file and suggests `centinela evidence read` so the user can
// inspect predecessor state before retrying.
//
// The acquire/release primitives live in the internal/filelock leaf, shared
// with internal/roadmap's roadmap.json lock — one flock implementation, so the
// two domains cannot drift apart on what "held" means.
func Lock(feature string, role Role) (func(), error) {
	hint := fmt.Sprintf("another agent is writing — try `centinela evidence read %s %s` first",
		feature, role)
	release, err := filelock.Acquire(lockPath(feature, role), LockTimeout, LockPollInterval, hint)
	if err != nil {
		return nil, fmt.Errorf("evidence %w", err)
	}
	return release, nil
}
