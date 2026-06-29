package cost

import (
	"encoding/json"
	"os"
)

// cursorPath stores the transcript read position so repeated capture fires add
// only the new tokens. It lives beside the telemetry log (local-only, gitignored
// telemetry dir). Path is recorded so a new session (different transcript)
// resets the offset to 0.
const cursorDir = ".workflow/telemetry"
const cursorFile = ".workflow/telemetry/cost-cursor.json"

// Cursor is the persisted transcript read position.
type Cursor struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
}

// LoadCursor returns the saved cursor, or a zero Cursor when absent/unreadable.
func LoadCursor() Cursor {
	data, err := os.ReadFile(cursorFile)
	if err != nil {
		return Cursor{}
	}
	var c Cursor
	if json.Unmarshal(data, &c) != nil {
		return Cursor{}
	}
	return c
}

// OffsetFor returns the saved offset only when it belongs to transcriptPath; a
// different (or empty) path means a new session, so reading starts at 0.
func (c Cursor) OffsetFor(transcriptPath string) int64 {
	if c.Path == transcriptPath {
		return c.Offset
	}
	return 0
}

// SaveCursor persists the new position, best-effort (errors are swallowed —
// cost capture must never fail the host command).
func SaveCursor(transcriptPath string, offset int64) {
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		return
	}
	data, err := json.Marshal(Cursor{Path: transcriptPath, Offset: offset})
	if err != nil {
		return
	}
	_ = os.WriteFile(cursorFile, data, 0o644)
}
