package audit

import (
	"sort"

	"github.com/samuelnp/centinela/internal/config"
)

// Diff partitions the current violations against a recorded baseline.
//   - New: present now, absent from baseline — blocking.
//   - Baselined: present in both — tolerated, still reported.
//   - Resolved: in baseline, gone now — prunable on the next record.
//
// All three lists are sorted by Hash for deterministic rendering.
type Diff struct {
	New       []Fingerprint
	Baselined []Fingerprint
	Resolved  []Fingerprint
}

// Ratchet compares the current full-scan Fail set against the baseline by Hash.
// Comparison is Hash-only, so a Raw churn (a line-count change) never moves a
// violation's identity — a baselined oversized file stays baselined after edits.
func Ratchet(cfg *config.Config, b Baseline) Diff {
	current := currentFingerprints(cfg)
	baseline := baselineFingerprints(b)

	var d Diff
	for hash, fp := range current {
		if _, ok := baseline[hash]; ok {
			d.Baselined = append(d.Baselined, fp)
		} else {
			d.New = append(d.New, fp)
		}
	}
	for hash, fp := range baseline {
		if _, ok := current[hash]; !ok {
			d.Resolved = append(d.Resolved, fp)
		}
	}
	sortByHash(d.New)
	sortByHash(d.Baselined)
	sortByHash(d.Resolved)
	return d
}

// HasNew reports whether the ratchet found any new (un-baselined) violation. The
// standalone audit command exits non-zero iff this is true.
func (d Diff) HasNew() bool { return len(d.New) > 0 }

// baselineFingerprints flattens a baseline to a hash-keyed lookup.
func baselineFingerprints(b Baseline) map[string]Fingerprint {
	out := make(map[string]Fingerprint)
	for _, e := range b.Gates {
		for _, fp := range e.Fingerprints {
			out[fp.Hash] = fp
		}
	}
	return out
}

func sortByHash(fps []Fingerprint) {
	sort.Slice(fps, func(i, j int) bool { return fps[i].Hash < fps[j].Hash })
}
