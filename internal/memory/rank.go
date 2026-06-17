package memory

// applyCaps trims ranked entries to the count and byte budgets (SC-08/11).
func applyCaps(entries []Entry, maxEntries, maxBytes int) []Entry {
	out := []Entry{}
	used := 0
	for _, e := range entries {
		if len(out) >= maxEntries {
			break
		}
		size := e.sizeBytes()
		if used+size > maxBytes {
			break
		}
		used += size
		out = append(out, e)
	}
	return out
}

// FeatureTags returns the distinct tags of the planning feature's own ledger
// entries, used as its tag profile for shared-tag relevance matching.
func FeatureTags(feature string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, e := range loadEntries() {
		if e.Feature != feature {
			continue
		}
		for _, t := range e.Tags {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it] = true
	}
	return m
}
