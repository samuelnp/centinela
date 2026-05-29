//go:build windows

package evidence

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// tryLockExclusive attempts a non-blocking exclusive lock over the first
// byte of the file via LockFileEx. It returns (true, nil) on success,
// (false, nil) when another holder currently owns the lock, and
// (false, err) on any real failure. Locking a fixed single-byte range
// gives the same advisory mutual exclusion the Unix flock path provides.
func tryLockExclusive(f *os.File) (bool, error) {
	ol := new(windows.Overlapped)
	err := windows.LockFileEx(windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, ol)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return false, nil
	}
	return false, err
}

// unlockFile releases the LockFileEx lock held on the first byte of f.
func unlockFile(f *os.File) error {
	ol := new(windows.Overlapped)
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ol)
}
