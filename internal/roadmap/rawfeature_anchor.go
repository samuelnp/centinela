package roadmap

import "fmt"

// anchorPos resolves a before/after anchor within the phase at phaseIdx to an
// insertion index. When both anchors are empty the feature appends at the end;
// --before yields the anchor's index, --after the index just past it. An anchor
// that names no feature in the phase is an error. Callers enforce that at most
// one of before/after is set; if both are, before takes precedence.
func (d *rawDoc) anchorPos(phaseIdx int, before, after string) (int, error) {
	p, err := d.decodePhase(phaseIdx)
	if err != nil {
		return 0, err
	}
	if before == "" && after == "" {
		return len(p.Features), nil
	}
	anchor := before
	if anchor == "" {
		anchor = after
	}
	for i, f := range p.Features {
		name, err := featureName(f)
		if err != nil {
			return 0, err
		}
		if name != anchor {
			continue
		}
		if before == "" {
			return i + 1, nil
		}
		return i, nil
	}
	return 0, fmt.Errorf("anchor feature %q not found in phase %q", anchor, p.Name)
}
