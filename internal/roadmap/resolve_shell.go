package roadmap

import (
	"encoding/json"
	"fmt"
)

// shellKey is every phase key EXCEPT features. features is the merged finding
// list, computed by mergeBacklog and written over whatever the sides carried.
const shellKey = "features"

// mergeShellKeys three-ways the Backlog phase's own keys — `note` above all,
// plus anything a future schema adds — against the base.
//
// Merging the shell "theirs first, ours over the top" silently dropped a
// one-sided incoming edit to `note`: any key both sides carried resolved to
// ours regardless of what the base said. `note` is a first-class Phase field
// rendered as the phase blockquote in ROADMAP.md, so the loss was real and it
// happened behind a green success line. Every other phase (mergePhase) and
// every top-level key (mergeRest) already three-way; this is the same rule,
// applied to the one object that was exempt.
func mergeShellKeys(name string, b, o, t *side) (map[string]json.RawMessage, error) {
	base, err := b.phaseObject(name)
	if err != nil {
		return nil, err
	}
	ours, err := o.phaseObject(name)
	if err != nil {
		return nil, err
	}
	theirs, err := t.phaseObject(name)
	if err != nil {
		return nil, err
	}
	shell := map[string]json.RawMessage{}
	for _, key := range mergedKeys(ours, theirs) {
		if key == shellKey {
			continue
		}
		got, ok := threeWay(base[key], ours[key], theirs[key])
		if !ok {
			return nil, &PhaseConflictError{Phase: name}
		}
		if got != nil {
			shell[key] = got
		}
	}
	return shell, nil
}

// phaseObject decodes one side's copy of the named phase into its keys, or an
// empty map when that side does not have the phase at all. The values come
// from already-compacted phase bytes, so they compare byte-for-byte.
func (s *side) phaseObject(name string) (map[string]json.RawMessage, error) {
	raw, ok := s.phase[name]
	if !ok {
		return map[string]json.RawMessage{}, nil
	}
	obj := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("invalid phase %q in %s of the conflict: %w", name, s.label, err)
	}
	return obj, nil
}
