package memory

import (
	"sort"

	"github.com/samuelnp/centinela/internal/config"
)

// Query carries the planning feature's relevance signals for recall ranking.
type Query struct {
	Feature      string
	Dependencies []string
	Tags         []string
}

// Recall returns the deterministically ranked, capped slice of ledger entries
// relevant to the planning feature. Disabled config or an empty ledger yields
// an empty result with no error (SC-10/12).
func Recall(q Query, cfg *config.Config) []Entry {
	if cfg == nil || !cfg.Memory.IsEnabled() {
		return nil
	}
	entries := loadEntries()
	if len(entries) == 0 {
		return nil
	}
	deps := toSet(q.Dependencies)
	tags := toSet(q.Tags)
	ranked := rank(entries, deps, tags)
	return applyCaps(ranked, config.NormalizeRecallMaxEntries(cfg.Memory.RecallMaxEntries),
		config.NormalizeRecallMaxBytes(cfg.Memory.RecallMaxBytes))
}

// rank orders entries by dependency-feature match, then shared tags, then
// recency (createdAt descending) as a deterministic tie-break (SC-09).
func rank(entries []Entry, deps, tags map[string]bool) []Entry {
	out := append([]Entry{}, entries...)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := score(out[i], deps, tags), score(out[j], deps, tags)
		if si != sj {
			return si > sj
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func score(e Entry, deps, tags map[string]bool) int {
	s := 0
	if deps[e.Feature] {
		s += 100
	}
	for _, t := range e.Tags {
		if tags[t] {
			s += 10
		}
	}
	return s
}
