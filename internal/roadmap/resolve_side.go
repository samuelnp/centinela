package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// side is one version of roadmap.json in a three-way merge, indexed by phase
// name. Phase bodies are kept COMPACTED so two sides that differ only in
// whitespace compare equal — formatting is never a semantic conflict.
type side struct {
	label string
	order []string
	phase map[string]json.RawMessage
	rest  map[string]json.RawMessage
}

// parseSide decodes one merge stage. An unparseable side is refused by name:
// guessing which half of a corrupt file was intended is exactly the silent
// data loss `resolve` exists to prevent.
func parseSide(label string, data []byte) (*side, error) {
	// An absent stage (the side added or deleted the whole file) reads as an
	// empty roadmap, so every phase on it counts as absent rather than as a
	// parse failure.
	if len(bytes.TrimSpace(data)) == 0 {
		data = []byte(`{"phases":[]}`)
	}
	doc, err := parseRawRoadmap(data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s of the conflict: %w", label, err)
	}
	s := &side{label: label, phase: map[string]json.RawMessage{}, rest: map[string]json.RawMessage{}}
	for key, raw := range doc.rest {
		compact, cerr := compactJSON(raw)
		if cerr != nil {
			return nil, fmt.Errorf("cannot parse %s of the conflict: %w", label, cerr)
		}
		s.rest[key] = compact
	}
	for i := range doc.phases {
		name, err := phaseName(doc.phases[i])
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s of the conflict: %w", label, err)
		}
		compact, err := compactJSON(doc.phases[i])
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s of the conflict: %w", label, err)
		}
		if _, seen := s.phase[name]; !seen {
			s.order = append(s.order, name)
		}
		s.phase[name] = compact
	}
	return s, nil
}

// compactJSON normalizes a raw JSON value to its whitespace-free form.
func compactJSON(raw json.RawMessage) (json.RawMessage, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return nil, err
	}
	return json.RawMessage(append([]byte(nil), buf.Bytes()...)), nil
}

// features decodes the feature array of the named phase, or nil when the phase
// is absent on this side.
func (s *side) features(name string) ([]json.RawMessage, error) {
	raw, ok := s.phase[name]
	if !ok {
		return nil, nil
	}
	var phase map[string]json.RawMessage
	if err := json.Unmarshal(raw, &phase); err != nil {
		return nil, err
	}
	f, ok := phase["features"]
	if !ok {
		return nil, nil
	}
	var feats []json.RawMessage
	if err := json.Unmarshal(f, &feats); err != nil {
		return nil, fmt.Errorf("invalid features array in %s of %q: %w", s.label, name, err)
	}
	return feats, nil
}
