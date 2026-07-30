package worktree

import "strings"

// specKey identifies a scenario by the file that declares it, so that only
// the SAME scenario in the SAME file is ever compared. Two different
// scenarios that happen to share a Given clause are normal Gherkin, not a
// conflict.
func specKey(r scenarioRecord) string {
	return r.File + "\x00" + r.Scenario
}

// indexByKey maps file+scenario to its first record.
func indexByKey(recs []scenarioRecord) map[string]scenarioRecord {
	idx := make(map[string]scenarioRecord, len(recs))
	for _, r := range recs {
		if _, dup := idx[specKey(r)]; !dup {
			idx[specKey(r)] = r
		}
	}
	return idx
}

// diverges reports whether two copies of the same scenario differ in the
// content this detector understands. Byte-identical copies never diverge.
func diverges(a, b scenarioRecord) bool {
	return a.Given != b.Given || a.Then != b.Then
}

// bothEdited reports whether the two worktree copies each moved away from the
// main baseline. Every worktree is a full checkout, so an untouched copy of
// main's spec is carried by every tree; only a copy that differs from the
// baseline represents work someone would lose.
func bothEdited(a, b scenarioRecord, baseline map[string]scenarioRecord) bool {
	base, ok := baseline[specKey(a)]
	return !ok || (diverges(a, base) && diverges(b, base))
}

// appendDivergences adds one conflict per (fileA, ownerA, fileB, ownerB,
// scenario) tuple, skipping tuples already recorded in seen.
func appendDivergences(dst []SpecConflict, seen map[string]bool, merging []scenarioRecord,
	other ownedSpecs, baseline map[string]scenarioRecord) []SpecConflict {
	idx := indexByKey(other.recs)
	for _, m := range merging {
		o, ok := idx[specKey(m)]
		if !ok || !diverges(m, o) || !bothEdited(m, o, baseline) {
			continue
		}
		key := strings.Join([]string{m.File, m.Owner, o.File, other.owner, m.Scenario}, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		dst = append(dst, SpecConflict{
			FeatureA: m.File, OwnerA: m.Owner,
			FeatureB: o.File, OwnerB: other.owner,
			Scenario: m.Scenario, Given: m.Given,
		})
	}
	return dst
}
