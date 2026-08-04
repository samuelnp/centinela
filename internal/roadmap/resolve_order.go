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
func earlier(a, b json.RawMessage) json.RawMessage {
	ka, kb := keyOf(a), keyOf(b)
	if ka.deferredAt == "" || (kb.deferredAt != "" && kb.deferredAt < ka.deferredAt) {
		return b
	}
	return a
}

// backlogPhase rebuilds the Backlog phase object around the merged findings,
// preserving whatever other phase keys (note, and anything unknown) the
// surviving side carried. Ours wins the shell; theirs is the fallback when the
// phase exists only on their side. At least one side always has it — the
// caller reached here by finding the phase name on one of them.
func backlogPhase(name string, o, t *side, kept []json.RawMessage) (json.RawMessage, error) {
	shell := map[string]json.RawMessage{}
	for _, s := range []*side{t, o} {
		if raw, ok := s.phase[name]; ok {
			if err := json.Unmarshal(raw, &shell); err != nil {
				return nil, err
			}
		}
	}
	if len(kept) == 0 {
		kept = []json.RawMessage{}
	}
	feats, err := json.Marshal(kept)
	if err != nil {
		return nil, err
	}
	shell["features"] = feats
	return json.Marshal(shell)
}
