// Package filelock is an advisory cross-process file lock: the one primitive
// two different domains need to serialize a read-modify-write of a shared file.
//
// It was extracted from internal/evidence, which has used it since the
// evidence CLI shipped, when internal/roadmap needed the same guarantee for
// roadmap.json. Two independent flock implementations of the same idea is
// exactly the drift a leaf exists to prevent, and neither domain may import the
// other. Standard library + syscall only.
//
// The lock is ADVISORY: it only excludes writers that also take it. It is held
// on a sibling .lock file, never on the document itself, so the document is
// never opened with a lock held.
package filelock

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultTimeout is the upper bound on how long Acquire waits. Short by
// design: a CLI invocation holds the lock for well under a millisecond, so a
// wait this long already means a genuinely concurrent writer.
const DefaultTimeout = 2 * time.Second

// DefaultPollInterval is the retry cadence inside the timeout window.
const DefaultPollInterval = 25 * time.Millisecond

// Acquire takes an exclusive advisory lock on path, creating it and its parent
// directory if needed, and returns a release function the caller MUST defer.
//
// It polls rather than blocking so a stuck holder surfaces as a timeout with a
// message naming the file instead of hanging the command forever. busyHint is
// appended to that timeout error; pass "" for none.
func Acquire(path string, timeout, poll time.Duration, busyHint string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("lock mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("lock open %s: %w", path, err)
	}
	deadline := time.Now().Add(timeout)
	for {
		locked, err := tryLockExclusive(f)
		if err != nil {
			f.Close() //nolint:errcheck
			return nil, fmt.Errorf("lock %s: %w", path, err)
		}
		if locked {
			return func() {
				_ = unlockFile(f)
				_ = f.Close()
			}, nil
		}
		if time.Now().After(deadline) {
			f.Close() //nolint:errcheck
			return nil, busyError(path, timeout, busyHint)
		}
		time.Sleep(poll)
	}
}

// busyError phrases the timeout so the operator knows who to look for.
func busyError(path string, timeout time.Duration, hint string) error {
	msg := fmt.Sprintf("lock busy after %s for %s; another process is writing", timeout, path)
	if hint != "" {
		msg += " — " + hint
	}
	return fmt.Errorf("%s", msg)
}
