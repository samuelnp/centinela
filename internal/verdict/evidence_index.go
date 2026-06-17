package verdict

import (
	"os"
	"sort"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// EvidenceIndex lists the on-disk role evidence for a feature, one entry per
// .workflow/<feature>-<role>.json file that exists. Entries are sorted by role
// name. Returns an empty (non-nil) slice when no evidence exists so the packet
// always emits a JSON array.
func EvidenceIndex(feature string) []EvidLine {
	out := make([]EvidLine, 0, len(evidence.AllRoles()))
	for _, role := range evidence.AllRoles() {
		path := orchestration.JSONPath(feature, role)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		ev, err := evidence.Read(feature, role)
		if err != nil || ev == nil {
			continue
		}
		out = append(out, EvidLine{
			Role:        string(role),
			Step:        ev.Step,
			Status:      ev.Status,
			HandoffTo:   ev.HandoffTo,
			GeneratedAt: ev.GeneratedAt,
			Path:        path,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Role < out[j].Role })
	return out
}
