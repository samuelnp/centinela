package evidence

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

// LockTimeout is the upper bound on how long Lock waits before giving up.
// Short by design — the typical CLI invocation holds the lock for a few
// hundred microseconds, so even one second is generous.
const LockTimeout = 2 * time.Second

// LockPollInterval is the retry cadence inside the LockTimeout window.
const LockPollInterval = 25 * time.Millisecond

// lockPath returns the .lock sibling file we lock for advisory mutual
// exclusion. Keeping it separate from the JSON means the JSON itself is
// never opened with a held lock.
func lockPath(feature string, role Role) string {
	return filepath.Join(workflow.WorkflowDir, fmt.Sprintf("%s-%s.lock", feature, role))
}

// Lock takes an advisory file lock on the (feature, role) pair and returns
// a release function that callers MUST defer. On timeout the error message
// names the file and suggests `centinela evidence read` so the user can
// inspect predecessor state before retrying. The OS-specific acquire/release
// primitives live in lock_unix.go and lock_windows.go.
func Lock(feature string, role Role) (func(), error) {
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		return nil, fmt.Errorf("evidence lock mkdir: %w", err)
	}
	path := lockPath(feature, role)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("evidence lock open %s: %w", path, err)
	}
	deadline := time.Now().Add(LockTimeout)
	for {
		locked, err := tryLockExclusive(f)
		if err != nil {
			f.Close() //nolint:errcheck
			return nil, fmt.Errorf("evidence lock %s: %w", path, err)
		}
		if locked {
			return func() {
				_ = unlockFile(f)
				_ = f.Close()
			}, nil
		}
		if time.Now().After(deadline) {
			f.Close() //nolint:errcheck
			return nil, fmt.Errorf("evidence lock busy after %s for %s; another agent is writing — try `centinela evidence read %s %s` first",
				LockTimeout, path, feature, role)
		}
		time.Sleep(LockPollInterval)
	}
}
