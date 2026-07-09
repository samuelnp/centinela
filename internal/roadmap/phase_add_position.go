package roadmap

import "fmt"

// insertPosition resolves where a new phase lands. With afterPhase set the phase
// lands immediately after that named phase (Backlog/Baseline allowed as a
// position anchor — the new phase is still a normal schedulable phase, not a
// member of them); an unknown anchor is refused. With no anchor it lands just
// before the Backlog phase, or last when no Backlog phase exists.
func (d *rawDoc) insertPosition(afterPhase string) (int, error) {
	if afterPhase != "" {
		idx, err := d.phaseIndexByName(afterPhase)
		if err != nil {
			return -1, err
		}
		if idx < 0 {
			return -1, fmt.Errorf("unknown phase %q for --after anchor", afterPhase)
		}
		return idx + 1, nil
	}
	backlogIdx, err := d.backlogPhaseIndex()
	if err != nil {
		return -1, err
	}
	if backlogIdx >= 0 {
		return backlogIdx, nil
	}
	return len(d.phases), nil
}
