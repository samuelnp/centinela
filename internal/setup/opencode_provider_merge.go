package setup

import "encoding/json"

// localProviderMarker tags a Centinela-managed provider block so a re-run can
// tell its own block (safe to update on a real change) from a same-key block a
// user hand-wrote (never overwritten). It sits beside npm/options/models; OpenCode
// reads only those keys, so the marker is inert at runtime.
const localProviderMarker = "centinela:managed-version=" + setupDocVersion

// mergeProvider adds or updates ONLY Centinela's managed local provider block
// under its provider key. nil → no-op (zero-config output unchanged). It never
// touches a foreign provider key, and never overwrites a same-key block that
// lacks the managed marker. Idempotent: it rewrites only on a real value diff.
func mergeProvider(raw map[string]json.RawMessage, lp *LocalProvider) bool {
	if lp == nil {
		return false
	}
	providers := map[string]json.RawMessage{}
	_ = json.Unmarshal(raw["provider"], &providers)
	key, block := buildLocalProvider(*lp)
	block["centinela"] = localProviderMarker
	want, _ := json.Marshal(block)
	if existing, ok := providers[key]; ok {
		if !isManagedProvider(existing) {
			return false
		}
		if string(normalizeBlock(existing)) == string(want) {
			return false
		}
	}
	providers[key] = want
	raw["provider"], _ = json.Marshal(providers)
	return true
}

// isManagedProvider reports whether a provider block carries Centinela's marker.
func isManagedProvider(block json.RawMessage) bool {
	var m map[string]json.RawMessage
	if json.Unmarshal(block, &m) != nil {
		return false
	}
	var marker string
	_ = json.Unmarshal(m["centinela"], &marker)
	return marker == localProviderMarker
}

// normalizeBlock canonicalizes a raw JSON block (sorted keys) for value compare.
func normalizeBlock(block json.RawMessage) []byte {
	var v any
	if json.Unmarshal(block, &v) != nil {
		return nil
	}
	out, _ := json.Marshal(v)
	return out
}
