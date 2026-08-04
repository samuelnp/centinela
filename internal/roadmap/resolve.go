package roadmap

import (
	"encoding/json"
	"fmt"
)

// PhaseConflictError reports a schedulable phase both sides changed. v1 refuses
// rather than guesses: silently unioning two real edits is how a finding gets
// lost, which is the failure this command exists to prevent.
type PhaseConflictError struct{ Phase string }

// Error names the phase the operator must resolve by hand.
func (e *PhaseConflictError) Error() string {
	return fmt.Sprintf("phase %q changed on both sides; resolve it by hand — "+
		"roadmap resolve only merges divergent Backlog findings", e.Phase)
}

// Merged is a resolved roadmap.json plus the arithmetic a human needs to trust
// it: how many findings survived, and how many each side contributed.
type Merged struct {
	Doc        []byte
	Kept       int
	FromOurs   int
	FromTheirs int
}

// Resolve performs a semantic three-way merge of a conflicted roadmap.json.
//
// Backlog findings are unioned by slug (see mergeBacklog). Every other phase,
// and every non-phase top-level key, takes whichever side differs from base;
// when BOTH differ it returns a *PhaseConflictError naming the phase and
// writes nothing, leaving the operator's conflict markers in place.
func Resolve(base, ours, theirs []byte) (Merged, error) {
	b, err := parseSide("the base", base)
	if err != nil {
		return Merged{}, err
	}
	o, err := parseSide("our side", ours)
	if err != nil {
		return Merged{}, err
	}
	t, err := parseSide("their side", theirs)
	if err != nil {
		return Merged{}, err
	}
	doc := &rawDoc{rest: map[string]json.RawMessage{}, dirty: map[int]string{}}
	out := Merged{}
	for _, name := range mergedOrder(o, t) {
		phase, err := mergePhase(name, b, o, t, &out)
		if err != nil {
			return Merged{}, err
		}
		if phase != nil {
			doc.phases = append(doc.phases, phase)
		}
	}
	if doc.rest, err = mergeRest(b, o, t); err != nil {
		return Merged{}, err
	}
	if out.Doc, err = doc.render(); err != nil {
		return Merged{}, err
	}
	return out, nil
}

// mergePhase resolves one phase: the Backlog semantically, anything else by
// taking the side that moved.
func mergePhase(name string, b, o, t *side, out *Merged) (json.RawMessage, error) {
	if isBacklogPhaseName(name) {
		return mergeBacklog(name, b, o, t, out)
	}
	got, ok := threeWay(b.phase[name], o.phase[name], t.phase[name])
	if !ok {
		return nil, &PhaseConflictError{Phase: name}
	}
	return got, nil
}
