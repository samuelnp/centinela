package roadmap

import "fmt"

// finalizeMutation re-decodes the mutated doc, re-validates dependency integrity
// (unknown-dep, cycle, archetype), and performs the single atomic write. Shared
// by edit/move/reorder so every mutation path enforces the same post-conditions
// before touching disk; a validation failure leaves roadmap.json byte-identical.
func finalizeMutation(path string, doc *rawDoc) error {
	typed, err := doc.toRoadmap()
	if err != nil {
		return err
	}
	if err := ValidateDependencies(typed); err != nil {
		return err
	}
	return writeRawRoadmap(path, doc)
}

// schedulablePhaseIndex returns the index of the named schedulable phase,
// refusing Backlog/Baseline (non-schedulable) and unknown names.
func (d *rawDoc) schedulablePhaseIndex(name string) (int, error) {
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return -1, err
		}
		if p.Name != name {
			continue
		}
		if isNonSchedulablePhase(p.Name) {
			return -1, fmt.Errorf("phase %q is non-schedulable; move/reorder targets a schedulable phase", name)
		}
		return i, nil
	}
	return -1, fmt.Errorf("unknown phase %q; known phases: %s", name, d.knownPhaseList())
}

// requireSchedulablePhaseIdx refuses a mutation whose phase at idx is Backlog or
// Baseline, whose entries are non-schedulable and must not be moved/reordered.
func (d *rawDoc) requireSchedulablePhaseIdx(idx int) error {
	p, err := d.decodePhase(idx)
	if err != nil {
		return err
	}
	if isNonSchedulablePhase(p.Name) {
		return fmt.Errorf("feature is in %q, a non-schedulable phase; move/reorder is refused", p.Name)
	}
	return nil
}
