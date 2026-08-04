package roadmap

import (
	"encoding/json"
	"fmt"
	"sort"
)

// indexFindings maps slug -> raw entry for one side's copy of the phase. A
// duplicate slug within one side keeps its first occurrence. An entry with no
// name is REFUSED rather than keyed on "": two nameless findings from opposite
// sides would otherwise dedupe into one, which is the silent data loss this
// command exists to prevent.
func indexFindings(s *side, phase string) (map[string]json.RawMessage, error) {
	feats, err := s.features(phase)
	if err != nil {
		return nil, err
	}
	idx := make(map[string]json.RawMessage, len(feats))
	for _, f := range feats {
		slug, err := featureName(f)
		if err != nil {
			return nil, err
		}
		if slug == "" {
			return nil, fmt.Errorf("a %s Backlog entry has no name: %s", s.label, f)
		}
		if _, seen := idx[slug]; !seen {
			idx[slug] = f
		}
	}
	return idx, nil
}

// sortedSlugs is the deterministic union of both sides' slugs.
func sortedSlugs(a, b map[string]json.RawMessage) []string {
	seen := map[string]bool{}
	var slugs []string
	for _, m := range []map[string]json.RawMessage{a, b} {
		for slug := range m {
			if !seen[slug] {
				seen[slug] = true
				slugs = append(slugs, slug)
			}
		}
	}
	sort.Strings(slugs)
	return slugs
}
