package worktree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PendingMarker records that a merge stalled and needs the Merge Steward.
// It is rewritten (never appended) every time a merge re-stalls so the
// latest reason always wins.
type PendingMarker struct {
	Feature         string   `json:"feature"`
	Reason          string   `json:"reason"`
	ConflictedPaths []string `json:"conflictedPaths"`
	WorktreePath    string   `json:"worktreePath"`
	GeneratedAt     string   `json:"generatedAt"`
}

// PendingPath returns the marker location for a feature relative to repo.
func PendingPath(repo, feature string) string {
	return filepath.Join(repo, ".workflow", fmt.Sprintf("%s-merge-pending.json", feature))
}

// WritePending rewrites the marker from a stalled merge outcome. It
// truncates any existing marker so the file is idempotent across reruns.
func WritePending(repo string, o MergeOutcome) error {
	m := PendingMarker{
		Feature:         o.Feature,
		Reason:          o.StewardReason(),
		ConflictedPaths: o.ConflictedPaths,
		WorktreePath:    Path(repo, o.Feature),
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	if err := os.MkdirAll(filepath.Join(repo, ".workflow"), 0o755); err != nil {
		return fmt.Errorf("pending marker: cannot create .workflow: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("pending marker: cannot encode: %w", err)
	}
	if err := os.WriteFile(PendingPath(repo, o.Feature), data, 0o644); err != nil {
		return fmt.Errorf("pending marker: cannot write: %w", err)
	}
	return nil
}

// LoadPending reads the marker. A missing file returns (nil, nil) so
// callers can treat absence as "nothing pending".
func LoadPending(repo, feature string) (*PendingMarker, error) {
	data, err := os.ReadFile(PendingPath(repo, feature))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pending marker: cannot read: %w", err)
	}
	var m PendingMarker
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("pending marker: corrupt json: %w", err)
	}
	return &m, nil
}

// Directive re-renders the dispatch directive from a stored marker so the
// UserPromptSubmit hook can re-emit it verbatim while the marker lives.
func (m *PendingMarker) Directive() string {
	o := MergeOutcome{Feature: m.Feature, ConflictedPaths: m.ConflictedPaths}
	switch m.Reason {
	case "git-text-conflict":
		o.TextConflict = true
	case "post-merge-validate-failed":
		o.ValidateFail = true
	}
	return o.StewardDirective()
}

// ClearPending removes the marker. Idempotent: absence is not an error.
func ClearPending(repo, feature string) error {
	err := os.Remove(PendingPath(repo, feature))
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("pending marker: cannot clear: %w", err)
}
