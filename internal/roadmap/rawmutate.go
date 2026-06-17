package roadmap

import "encoding/json"

// rawPhase is a phase decoded for mutation: its name plus features as raw
// per-entry bytes so untouched entries round-trip unchanged.
type rawPhase struct {
	Name     string            `json:"name"`
	Features []json.RawMessage `json:"features"`
}

// phaseFeatureNames returns every feature name across all phases (raw scan).
func (d *rawDoc) phaseFeatureNames() (map[string]string, error) {
	out := map[string]string{} // feature name -> phase name
	for i := range d.phases {
		p, err := d.decodePhase(i)
		if err != nil {
			return nil, err
		}
		for _, f := range p.Features {
			name, err := featureName(f)
			if err != nil {
				return nil, err
			}
			out[name] = p.Name
		}
	}
	return out, nil
}

func (d *rawDoc) decodePhase(i int) (*rawPhase, error) {
	var p rawPhase
	if err := json.Unmarshal(d.phaseBytes(i), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// setPhase re-encodes a mutated phase and marks it dirty.
func (d *rawDoc) setPhase(i int, p *rawPhase) error {
	body, err := encodePhase(p)
	if err != nil {
		return err
	}
	d.dirty[i] = string(body)
	return nil
}

// appendBacklog ensures a Backlog phase exists (as the last phase) and appends
// the given feature entry to it. Returns the resulting phase index.
func (d *rawDoc) appendBacklog(entry json.RawMessage) error {
	idx, err := d.backlogPhaseIndex()
	if err != nil {
		return err
	}
	if idx < 0 {
		p := &rawPhase{Name: BacklogPhaseName, Features: []json.RawMessage{entry}}
		body, err := encodePhase(p)
		if err != nil {
			return err
		}
		d.dirty[len(d.phases)] = string(body)
		d.phases = append(d.phases, json.RawMessage(body))
		return nil
	}
	p, err := d.decodePhase(idx)
	if err != nil {
		return err
	}
	p.Features = append(p.Features, entry)
	return d.setPhase(idx, p)
}

func featureName(f json.RawMessage) (string, error) {
	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(f, &obj); err != nil {
		return "", err
	}
	return obj.Name, nil
}

// encodePhase marshals a phase compactly (no HTML escaping); render() re-indents
// it, emitting the features array one object per line for merge-friendly diffs.
func encodePhase(p *rawPhase) (json.RawMessage, error) {
	return compactBytes(p)
}
