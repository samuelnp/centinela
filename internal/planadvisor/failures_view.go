package planadvisor

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/insights"
)

// failureSummary renders the ranked gate failures as "gate (×N)" entries joined
// by commas, preserving the insights rank order.
func failureSummary(fs []insights.Count) string {
	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		parts = append(parts, fmt.Sprintf("%s (×%d)", f.Key, f.Count))
	}
	return strings.Join(parts, ", ")
}
