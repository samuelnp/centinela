//go:build !windows

package evidence

import (
	"errors"
	"os"
	"syscall"
)

// tryLockExclusive attempts a non-blocking exclusive advisory lock via
// flock(2). It returns (true, nil) on success, (false, nil) when another
// holder currently owns the lock, and (false, err) on any real failure.
func tryLockExclusive(f *os.File) (bool, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	return false, err
}

// unlockFile releases the flock held on f.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
