package roadmap

import "encoding/json"

// applyReorder removes the feature from its source phase then re-inserts it at the
// anchor's resolved position. The anchor's phase must be schedulable; resolving
// the anchor after the removal keeps indices correct when source and anchor share
// a phase.
func (d *rawDoc) applyReorder(req ReorderRequest, srcIdx int, anchor string, entry json.RawMessage) error {
	_, targetIdx, _, err := d.findFeature(anchor)
	if err != nil {
		return err
	}
	if err := d.requireSchedulablePhaseIdx(targetIdx); err != nil {
		return err
	}
	if err := d.removeFeatureAt(srcIdx, req.Slug); err != nil {
		return err
	}
	pos, err := d.anchorPos(targetIdx, req.BeforeAnchor, req.AfterAnchor)
	if err != nil {
		return err
	}
	return d.insertFeatureAt(targetIdx, pos, entry)
}

// phaseOrder returns each phase's ordered feature-name list. Reorder compares the
// order before and after the mutation to detect an order-preserving no-op, which
// must leave roadmap.json byte-identical (so it is not rewritten at all).
func (d *rawDoc) phaseOrder() ([][]string, error) {
	out := make([][]string, len(d.phases))
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return nil, err
		}
		names := make([]string, len(p.Features))
		for j, f := range p.Features {
			name, err := featureName(f)
			if err != nil {
				return nil, err
			}
			names[j] = name
		}
		out[i] = names
	}
	return out, nil
}

// sameOrder reports whether two phase-order snapshots are element-wise identical.
func sameOrder(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
