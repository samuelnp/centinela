package roadmap

import (
	"encoding/json"
	"fmt"
)

// findFeature locates a feature by slug anywhere in the doc, returning its raw
// bytes plus the owning phase and feature indices. Errors when absent.
func (d *rawDoc) findFeature(slug string) (json.RawMessage, int, int, error) {
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return nil, -1, -1, err
		}
		for j, f := range p.Features {
			name, err := featureName(f)
			if err != nil {
				return nil, -1, -1, err
			}
			if name == slug {
				return f, i, j, nil
			}
		}
	}
	return nil, -1, -1, fmt.Errorf("feature %q not found in roadmap", slug)
}

// featurePhase returns the name of the phase that owns slug, or an error.
func (d *rawDoc) featurePhase(slug string) (string, error) {
	_, phaseIdx, _, err := d.findFeature(slug)
	if err != nil {
		return "", err
	}
	p, err := d.decodePhase(phaseIdx)
	if err != nil {
		return "", err
	}
	return p.Name, nil
}
