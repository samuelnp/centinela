package roadmap

import (
	"encoding/json"
	"sort"
)

// sortedKeys returns the map's keys in stable ascending order so map-backed
// JSON re-emission is deterministic (Go map iteration is randomized, which
// would otherwise churn untouched keys into spurious git diffs between runs).
func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
