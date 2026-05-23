package roadmapcheckpoint

import (
	"os"
	"time"
)

// osFS is the production FS implementation backed by the os package. It is
// unexported; callers obtain it through NewOSFS so the concrete type stays
// internal and the FS contract is the only public surface.
type osFS struct{}

// NewOSFS returns an FS that reads the real filesystem via the os package.
// Use this in cmd/centinela to drive Decide against on-disk state.
func NewOSFS() FS {
	return osFS{}
}

// Stat returns the modification time for path. ok is false when the path is
// missing or otherwise unstattable; callers treat that as "absent".
func (osFS) Stat(path string) (time.Time, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, false
	}
	return info.ModTime(), true
}

// ReadFile returns the bytes at path. ok is false when the file is missing
// or unreadable, matching the FS contract used by Decide.
func (osFS) ReadFile(path string) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Exists reports whether path exists on disk.
func (osFS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
