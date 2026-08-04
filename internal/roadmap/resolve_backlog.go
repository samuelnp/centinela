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
// the caller prints the per-side arithmetic so a human can check it (R5).
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
		entry, ok := survivor(slug, baseIdx, ourIdx, theirIdx)
		if !ok {
			continue
		}
		kept = append(kept, entry)
		countContribution(slug, baseIdx, ourIdx, theirIdx, out)
	}
	sort.SliceStable(kept, func(i, j int) bool { return byDeferredAtThenName(kept[i], kept[j]) })
	out.Kept = len(kept)
	return backlogPhase(name, o, t, kept)
}

// survivor applies the presence rules and picks the entry to keep.
func survivor(slug string, base, ours, theirs map[string]json.RawMessage) (json.RawMessage, bool) {
	mine, hasMine := ours[slug]
	yours, hasYours := theirs[slug]
	_, hadBase := base[slug]
	if hasMine && hasYours {
		return earlier(mine, yours), true
	}
	if hasMine {
		return mine, !hadBase // present on base and gone from theirs => deleted
	}
	// Callers only ask about slugs at least one side has, so this is theirs.
	return yours, hasYours && !hadBase
}

// countContribution attributes a kept finding that the base did not have.
// A slug both sides added counts once, as ours, so the arithmetic sums.
func countContribution(slug string, base, ours, theirs map[string]json.RawMessage, out *Merged) {
	if _, hadBase := base[slug]; hadBase {
		return
	}
	if _, ok := ours[slug]; ok {
		out.FromOurs++
		return
	}
	out.FromTheirs++
}
