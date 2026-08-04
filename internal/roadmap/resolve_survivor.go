package roadmap

import (
	"bytes"
	"encoding/json"
)

// survivor decides one slug. Presence alone is NOT enough, and neither is
// deferredAt: an entry the base had is resolved against the base first, exactly
// like a phase, so a one-sided EDIT wins over an untouched side whichever side
// made it. Deciding by presence threw away a modify/delete edit; deciding by
// deferredAt threw away an incoming edit, because an edit does not move the
// timestamp and the tie silently resolved to ours.
func survivor(slug string, base, ours, theirs map[string]json.RawMessage) (json.RawMessage, bool, error) {
	mine, hasMine := ours[slug]
	yours, hasYours := theirs[slug]
	was, hadBase := base[slug]
	if hasMine && hasYours {
		if !hadBase {
			// Both sides independently deferred the same slug: no base to
			// three-way against, so keep the first capture.
			return earlier(mine, yours), true, nil
		}
		got, ok := threeWay(was, mine, yours)
		if !ok {
			return nil, false, &FindingConflictError{Slug: slug, Detail: detailBothEdited}
		}
		return got, true, nil
	}
	if !hadBase {
		// Added on exactly one side (or on neither): keep whichever has it.
		return oneSided(mine, hasMine, yours, hasYours)
	}
	// The base had it and exactly one side still does: the deletion wins only
	// over a side that did not touch the entry.
	kept, present := mine, hasMine
	if !hasMine {
		kept, present = yours, hasYours
	}
	if present && !bytes.Equal(kept, was) {
		return nil, false, &FindingConflictError{Slug: slug, Detail: detailModifyDelete}
	}
	return nil, false, nil
}

// oneSided returns whichever side holds an entry the base never had.
func oneSided(mine json.RawMessage, hasMine bool, yours json.RawMessage, hasYours bool) (json.RawMessage, bool, error) {
	if hasMine {
		return mine, true, nil
	}
	return yours, hasYours, nil
}

// countContribution credits the surviving ENTRY to the side whose bytes it
// actually is, so the summary a human is told to check (R5) is true and adds up.
//
// The old rule counted by presence of the SLUG and skipped anything the base
// already had, which produced two lies at once: the counts never summed to
// Kept, and a slug both sides deferred was credited to ours even when the
// earlier capture kept was theirs. Comparing bytes cannot misattribute.
func countContribution(entry json.RawMessage, slug string, base, ours map[string]json.RawMessage, out *Merged) {
	if was, hadBase := base[slug]; hadBase && bytes.Equal(entry, was) {
		out.FromBase++ // survived unchanged: neither side contributed it
		return
	}
	if mine, hasMine := ours[slug]; hasMine && bytes.Equal(entry, mine) {
		out.FromOurs++
		return
	}
	out.FromTheirs++
}
