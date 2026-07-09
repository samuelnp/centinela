package roadmap

import "fmt"

// PhaseRemove deletes a phase via raw-preserving read-modify-write. An unknown
// or reserved (Backlog/Baseline) phase is refused. An empty phase is removed
// directly. A non-empty phase is refused (naming the feature count) unless force
// is set, in which case the phase, its features, and those features' analysis and
// quality entries are removed together after a pre-write dependency re-validation
// that refuses the whole op (byte-identical) if a surviving feature would be left
// depending on a removed one. A rejected remove writes nothing.
func PhaseRemove(path, name string, force bool) error {
	if isNonSchedulablePhase(name) {
		return fmt.Errorf("%q is a reserved phase name; managed via defer/promote", name)
	}
	doc, err := readRawRoadmap(path)
	if err != nil {
		return err
	}
	idx, err := doc.phaseIndexByName(name)
	if err != nil {
		return err
	}
	if idx < 0 {
		return fmt.Errorf("phase %q not found", name)
	}
	p, err := doc.decodePhase(idx)
	if err != nil {
		return err
	}
	if len(p.Features) == 0 {
		if err := doc.removePhaseAt(idx); err != nil {
			return err
		}
		return finalizeMutation(path, doc)
	}
	if !force {
		return fmt.Errorf(
			"phase %q contains %d features; pass --force to remove the phase and its features",
			name, len(p.Features))
	}
	return doc.forceRemovePhase(path, idx, p)
}
