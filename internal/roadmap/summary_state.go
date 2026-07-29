package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Digest state I/O lives beside summary_digest.go: both halves fail open, but
// only this half touches the filesystem.

// LoadSummaryState never returns an error: absent, corrupt, truncated or
// unreadable state degrades to the zero value, which forces a render (AC17).
func LoadSummaryState(path string) SummaryState {
	data, err := os.ReadFile(path)
	if err != nil {
		return SummaryState{}
	}
	var s SummaryState
	if err := json.Unmarshal(data, &s); err != nil {
		return SummaryState{}
	}
	return s
}

// SaveSummaryState writes atomically (temp file + rename in the same directory)
// so two concurrent hook processes can at worst cost one extra render, never a
// partial-write corrupt state. Callers ignore the error — an unwritable
// .workflow/ must never surface to the host session.
func SaveSummaryState(path string, s SummaryState) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".roadmap-digest-*")
	if err != nil {
		return err
	}
	if err := writeAndRename(tmp, path, data); err != nil {
		os.Remove(tmp.Name()) //nolint:errcheck // best-effort cleanup of the temp file
		return err
	}
	return nil
}

// writeAndRename fills the temp file and moves it into place. Every failure
// leaves the caller responsible for removing the temp file, so no partially
// written state can ever be observed at path.
func writeAndRename(tmp *os.File, path string, data []byte) error {
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck // the write error is the one that matters
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
