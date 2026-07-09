package roadmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PhaseAdd inserts a new empty phase named name (with an optional note) via
// raw-preserving read-modify-write. With afterPhase set the phase lands just
// after that named phase; otherwise it lands just before the Backlog phase, or
// last when there is none. Reserved (Backlog/Baseline) names, duplicates, empty
// names, and an unknown --after anchor are refused, each leaving the file
// byte-identical. A single atomic write follows a post-insert dependency check.
func PhaseAdd(path, name, note, afterPhase string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("phase name is required")
	}
	if isNonSchedulablePhase(name) {
		return fmt.Errorf("%q is a reserved phase name; managed via defer/promote", name)
	}
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	if idx, err := doc.phaseIndexByName(name); err != nil {
		return err
	} else if idx >= 0 {
		return fmt.Errorf("phase %q already exists", name)
	}
	pos, err := doc.insertPosition(afterPhase)
	if err != nil {
		return err
	}
	entry, err := compactBytes(&rawPhase{Name: name, Note: note, Features: []json.RawMessage{}})
	if err != nil {
		return err
	}
	if err := doc.insertPhaseAt(pos, entry); err != nil {
		return err
	}
	return finalizeMutation(path, doc)
}
