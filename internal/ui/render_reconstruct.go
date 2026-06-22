package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/reconstruct"
)

// RenderReconstructionSummary renders the reconstruct command's stdout summary:
// targets selected, files written, files skipped (with their slugs), and the
// total "# TODO: confirm" markers across the corpus. Presentation only — it
// makes no decisions and performs no I/O.
func RenderReconstructionSummary(r reconstruct.Reconstruction) string {
	var b strings.Builder
	fmt.Fprintf(&b, "targets selected: %d\n", len(r.Targets))
	fmt.Fprintf(&b, "files written: %d\n", len(r.Written))
	fmt.Fprintf(&b, "files skipped: %d\n", len(r.Skipped))
	for _, s := range r.Skipped {
		b.WriteString("  - skipped (hand-authored spec exists): " + s + "\n")
	}
	fmt.Fprintf(&b, "TODO confirm markers: %d", r.TodoCount)
	return b.String()
}
