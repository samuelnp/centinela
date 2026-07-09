package roadmap

import (
	"fmt"
	"strings"
)

// PhaseRename renames a phase in place via raw-preserving read-modify-write. It
// refuses an unknown old name, an empty new name, a new name that collides with
// an existing phase, and either side being a reserved Backlog/Baseline name. A
// same-name request is a no-op (byte-identical, no write). The phase's features
// and every other phase round-trip byte-identically; a rejected rename writes
// nothing.
func PhaseRename(path, oldName, newName string) error {
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("phase name is required")
	}
	if isNonSchedulablePhase(oldName) || isNonSchedulablePhase(newName) {
		return fmt.Errorf("%q/%q is a reserved phase name; managed via defer/promote", oldName, newName)
	}
	if oldName == newName {
		return nil // no-op, byte-identical
	}
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	idx, err := doc.phaseIndexByName(oldName)
	if err != nil {
		return err
	}
	if idx < 0 {
		return fmt.Errorf("phase %q not found", oldName)
	}
	if collision, err := doc.phaseIndexByName(newName); err != nil {
		return err
	} else if collision >= 0 {
		return fmt.Errorf("phase %q already exists", newName)
	}
	if err := doc.renamePhaseAt(idx, newName); err != nil {
		return err
	}
	return finalizeMutation(path, doc)
}
