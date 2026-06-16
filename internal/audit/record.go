package audit

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Record runs every participating gate in full-repo scan and assembles a fresh,
// sorted baseline of the current Fail violations. It records the current set
// verbatim — resolved violations are never re-added, so the ratchet only ever
// tightens. The on-disk file is fully replaced by Save, not merged.
func Record(cfg *config.Config) Baseline {
	entries := currentEntries(cfg)
	b := Baseline{Scheme: fingerprintScheme, Version: 1, Gates: entries}
	sortBaseline(&b)
	return b
}

// currentEntries fingerprints the current full-scan Fail results of every
// participating gate. Only Status==Fail results yield baselineable violations;
// Warn/Skip/Pass produce none (this excludes import_graph's non-failing
// "unmapped packages" warning and any skipped gate). Gates with no surviving
// fingerprints are omitted.
func currentEntries(cfg *config.Config) []GateEntry {
	participating := participatingGates(cfg)
	results := gates.RunWithFilter(cfg, nil)
	var entries []GateEntry
	for _, r := range results {
		if r.Status != gates.Fail || !participating[r.Name] {
			continue
		}
		fps := Compute(r.Name, r.Details)
		if len(fps) == 0 {
			continue
		}
		entries = append(entries, GateEntry{Gate: r.Name, Fingerprints: fps})
	}
	return entries
}

// currentFingerprints flattens currentEntries to a hash-keyed lookup for the
// ratchet's set comparison.
func currentFingerprints(cfg *config.Config) map[string]Fingerprint {
	out := make(map[string]Fingerprint)
	for _, e := range currentEntries(cfg) {
		for _, fp := range e.Fingerprints {
			out[fp.Hash] = fp
		}
	}
	return out
}
