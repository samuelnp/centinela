package planadvisor

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/memory"
)

// recalledMemory returns one-line summaries of the ledger slice relevant to the
// planning feature. Empty ledger or disabled config yields nil (SC-10/12).
func recalledMemory(feature string, deps []string, cfg *config.Config) []string {
	q := memory.Query{
		Feature:      feature,
		Dependencies: deps,
		Tags:         memory.FeatureTags(feature),
	}
	entries := memory.Recall(q, cfg)
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Feature+" ["+e.Type+"]: "+e.Title)
	}
	return out
}
