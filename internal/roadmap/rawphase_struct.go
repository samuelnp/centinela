package roadmap

import (
	"encoding/json"
	"fmt"
)

// insertPhaseAt splices a new phase (raw bytes) into the phases slice at pos and
// reindexes the dirty map: every entry keyed >= pos shifts +1, because those
// phases now sit one slot later. The inserted phase is marked dirty at pos so it
// renders via renderDirtyPhase (features one object per line). 0 <= pos <= len.
func (d *rawDoc) insertPhaseAt(pos int, raw json.RawMessage) error {
	if pos < 0 || pos > len(d.phases) {
		return fmt.Errorf("insert position %d out of range", pos)
	}
	shifted := make(map[int]string, len(d.dirty))
	for k, v := range d.dirty {
		if k >= pos {
			shifted[k+1] = v
			continue
		}
		shifted[k] = v
	}
	d.phases = append(d.phases, nil)
	copy(d.phases[pos+1:], d.phases[pos:])
	d.phases[pos] = raw
	shifted[pos] = string(raw)
	d.dirty = shifted
	return nil
}

// removePhaseAt drops the phase at idx and reindexes the dirty map: the entry at
// idx is dropped and every entry keyed > idx shifts -1, because those phases now
// sit one slot earlier. An off-by-one here silently corrupts an unrelated phase,
// so the reindex lives ONLY here and in insertPhaseAt.
func (d *rawDoc) removePhaseAt(idx int) error {
	if idx < 0 || idx >= len(d.phases) {
		return fmt.Errorf("remove index %d out of range", idx)
	}
	shifted := make(map[int]string, len(d.dirty))
	for k, v := range d.dirty {
		switch {
		case k == idx:
			// dropped along with the phase
		case k > idx:
			shifted[k-1] = v
		default:
			shifted[k] = v
		}
	}
	d.phases = append(d.phases[:idx], d.phases[idx+1:]...)
	d.dirty = shifted
	return nil
}

// renamePhaseAt sets the name of the phase at idx and marks it dirty via setPhase
// (no structural shift, so the dirty map keeps its indices). Features round-trip
// verbatim through decodePhase/setPhase.
func (d *rawDoc) renamePhaseAt(idx int, newName string) error {
	p, err := d.decodePhase(idx)
	if err != nil {
		return err
	}
	p.Name = newName
	return d.setPhase(idx, p)
}

// phaseIndexByName returns the index of the phase named name (exact match), or
// -1 when absent. Reads through phaseBytes so a dirtied phase is still found.
func (d *rawDoc) phaseIndexByName(name string) (int, error) {
	for i := range d.phases {
		n, err := phaseName(d.phaseBytes(i))
		if err != nil {
			return -1, err
		}
		if n == name {
			return i, nil
		}
	}
	return -1, nil
}
