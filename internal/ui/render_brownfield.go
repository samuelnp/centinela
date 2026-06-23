package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/brownmap"
)

// RenderBrownfieldSummary renders the `centinela roadmap brownfield` stdout
// summary: the Baseline entry count, the gap feature count, and the draft path
// written. When there are zero gaps it appends a hint that the operator may
// supply --goal to add net-new work. Presentation only — it makes no decisions
// and performs no I/O.
func RenderBrownfieldSummary(p brownmap.Plan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "baseline entries: %d\n", p.BaselineCount)
	fmt.Fprintf(&b, "gaps: %d\n", p.GapCount)
	fmt.Fprintf(&b, "draft written: %s", p.DraftPath)
	if p.GapCount == 0 {
		b.WriteString("\nno gaps — supply --goal to add net-new work")
	}
	return b.String()
}
