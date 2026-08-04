package roadmap

import (
	"encoding/json"
	"sort"
)

// mergeBacklog unions the Backlog findings of both sides by slug.
//
// The rules, in the order they matter: a slug on BOTH sides survives once
// (keeping the entry with the EARLIER deferredAt, so provenance is the first
// capture); a slug added on one side survives; a slug the base had and exactly
// one side deleted stays deleted (promote/remove wins over an untouched side);
// a slug both sides deleted stays deleted. Nothing is ever silently dropped —
// the caller prints the per-side arithmetic so a human can check it (R5), and
// each survivor is credited to the side whose BYTES it is, so Kept reconciles
// as FromBase + FromOurs + FromTheirs.
func mergeBacklog(name string, b, o, t *side, out *Merged) (json.RawMessage, error) {
	baseIdx, err := indexFindings(b, name)
	if err != nil {
		return nil, err
	}
	ourIdx, err := indexFindings(o, name)
	if err != nil {
		return nil, err
	}
	theirIdx, err := indexFindings(t, name)
	if err != nil {
		return nil, err
	}
	var kept []json.RawMessage
	for _, slug := range sortedSlugs(ourIdx, theirIdx) {
		entry, ok, err := survivor(slug, baseIdx, ourIdx, theirIdx)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		kept = append(kept, entry)
		countContribution(entry, slug, baseIdx, ourIdx, out)
	}
	sort.SliceStable(kept, func(i, j int) bool { return byDeferredAtThenName(kept[i], kept[j]) })
	out.Kept = len(kept)
	return backlogPhase(name, b, o, t, kept)
}
