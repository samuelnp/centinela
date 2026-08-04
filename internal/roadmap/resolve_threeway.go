package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// mergedOrder lists every phase name once: ours' declared order first, then
// any phase only their side knows about, in their order.
func mergedOrder(o, t *side) []string {
	seen := map[string]bool{}
	order := make([]string, 0, len(o.order)+len(t.order))
	for _, name := range append(append([]string{}, o.order...), t.order...) {
		if !seen[name] {
			seen[name] = true
			order = append(order, name)
		}
	}
	return order
}

// threeWay picks the side that moved. Agreement wins; one-sided change wins
// (including a one-sided DELETE, where the moved side is nil); both-sides
// divergence is refused. Reported ok=false is always a real conflict.
func threeWay(base, ours, theirs json.RawMessage) (json.RawMessage, bool) {
	switch {
	case bytes.Equal(ours, theirs):
		return ours, true
	case bytes.Equal(ours, base):
		return theirs, true
	case bytes.Equal(theirs, base):
		return ours, true
	}
	return nil, false
}

// mergeRest three-ways the non-phase top-level keys (intro and anything a
// future schema adds), so a one-sided intro edit survives the merge.
func mergeRest(b, o, t *side) (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	for _, key := range mergedKeys(o.rest, t.rest) {
		got, ok := threeWay(b.rest[key], o.rest[key], t.rest[key])
		if !ok {
			return nil, fmt.Errorf("top-level key %q changed on both sides; resolve it by hand", key)
		}
		if got != nil {
			out[key] = got
		}
	}
	return out, nil
}

// mergedKeys is the deduplicated union of two key sets.
func mergedKeys(a, b map[string]json.RawMessage) []string {
	seen := map[string]bool{}
	var keys []string
	for _, m := range []map[string]json.RawMessage{a, b} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	return keys
}
