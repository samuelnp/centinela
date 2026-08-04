package roadmap

import (
	"encoding/json"
)

// findingKey is the pair the merged Backlog is ordered by.
type findingKey struct {
	deferredAt string
	name       string
}

// keyOf decodes the ordering fields of a raw finding. An entry that cannot be
// decoded sorts as an empty key — first, where a human will see it.
func keyOf(raw json.RawMessage) findingKey {
	var f struct {
		Name       string `json:"name"`
		DeferredAt string `json:"deferredAt"`
	}
	_ = json.Unmarshal(raw, &f)
	return findingKey{deferredAt: f.DeferredAt, name: f.Name}
}

// byDeferredAtThenName orders the merged Backlog: RFC 3339 sorts correctly as
// a string, and an absent deferredAt sorts first so it stays visible.
func byDeferredAtThenName(a, b json.RawMessage) bool {
	ka, kb := keyOf(a), keyOf(b)
	if ka.deferredAt != kb.deferredAt {
		return ka.deferredAt < kb.deferredAt
	}
	return ka.name < kb.name
}

// earlier returns whichever of two entries for the SAME slug was deferred
// first, so a dedupe keeps the original capture rather than the replay.
//
// It decides ONLY the case the base never had — both sides independently
// deferring one slug. It must never arbitrate an edit: deferredAt does not move
// when a finding is edited, so the comparison would tie and silently return
// ours. Anything the base had is three-wayed against the base in survivor.
func earlier(a, b json.RawMessage) json.RawMessage {
	ka, kb := keyOf(a), keyOf(b)
	if ka.deferredAt == "" || (kb.deferredAt != "" && kb.deferredAt < ka.deferredAt) {
		return b
	}
	return a
}

// backlogPhase rebuilds the Backlog phase object around the merged findings.
// The phase's own keys (note, and anything unknown) are three-wayed against the
// base like every other phase — see mergeShellKeys — and `features` is then
// written from the merged finding list.
func backlogPhase(name string, b, o, t *side, kept []json.RawMessage) (json.RawMessage, error) {
	shell, err := mergeShellKeys(name, b, o, t)
	if err != nil {
		return nil, err
	}
	if len(kept) == 0 {
		kept = []json.RawMessage{}
	}
	feats, err := json.Marshal(kept)
	if err != nil {
		return nil, err
	}
	shell[shellKey] = feats
	return json.Marshal(shell)
}
