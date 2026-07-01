package roadmap

import (
	"encoding/json"
	"fmt"
)

// appendFeatureToPhase appends a fully-formed feature entry to the named
// schedulable phase (Backlog and Baseline are refused as "unknown phase" so
// drafts live only in schedulable phases). The entry's own name is used for the
// per-phase duplicate guard; cross-phase collisions are caught by the caller.
func (d *rawDoc) appendFeatureToPhase(target string, entry json.RawMessage) error {
	name, err := featureName(entry)
	if err != nil {
		return err
	}
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return err
		}
		if isNonSchedulablePhase(p.Name) || p.Name != target {
			continue
		}
		for _, f := range p.Features {
			if n, _ := featureName(f); n == name {
				return fmt.Errorf("%q already exists in phase %q; refusing to add a duplicate", name, target)
			}
		}
		p.Features = append(p.Features, entry)
		return d.setPhase(i, p)
	}
	return fmt.Errorf("unknown phase %q; known phases: %s", target, d.knownPhaseList())
}

// removeFeatureAt drops the named feature from the phase at phaseIdx. Untouched
// entries round-trip byte-identically; the phase is left with an empty features
// array when its last feature is removed (phase removal is a successor feature).
func (d *rawDoc) removeFeatureAt(phaseIdx int, slug string) error {
	p, err := d.decodePhase(phaseIdx)
	if err != nil {
		return err
	}
	kept := p.Features[:0:0]
	for _, f := range p.Features {
		if name, _ := featureName(f); name != slug {
			kept = append(kept, f)
		}
	}
	p.Features = kept
	return d.setPhase(phaseIdx, p)
}

// replaceFeatureAt swaps the feature at (phaseIdx, featIdx) for entry, marking
// the phase dirty. Used by the in-place draft finalize to clear the draft flag.
func (d *rawDoc) replaceFeatureAt(phaseIdx, featIdx int, entry json.RawMessage) error {
	p, err := d.decodePhase(phaseIdx)
	if err != nil {
		return err
	}
	if featIdx < 0 || featIdx >= len(p.Features) {
		return fmt.Errorf("feature index %d out of range in phase %d", featIdx, phaseIdx)
	}
	p.Features[featIdx] = entry
	return d.setPhase(phaseIdx, p)
}
