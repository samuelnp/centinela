package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/synthesize"
)

// RenderInferenceSummary renders the inferred archetype, confidence, ambiguity
// note, and the winning rationale for the synthesize command output.
func RenderInferenceSummary(inf synthesize.Inference) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("inferred archetype: %s (confidence: %s)\n", inf.Best, inf.Confidence))
	if inf.Ambiguous {
		b.WriteString("note: ambiguous — top archetypes scored within the tie margin; confirm the choice\n")
	}
	reasons := inf.Reasons()
	if len(reasons) == 0 {
		b.WriteString("rationale: (none — no archetype signals matched; defaulted)\n")
	} else {
		b.WriteString("rationale:\n")
		for _, r := range reasons {
			b.WriteString("  - " + r + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
