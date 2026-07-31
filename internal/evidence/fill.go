package evidence

import "github.com/samuelnp/centinela/internal/orchestration"

// FillMarker re-exports the canonical substance-slot template so every existing
// caller in this package keeps its spelling. The constant itself lives in
// internal/orchestration because the workflow gate that rejects an unfilled slot
// must reach it too and cannot import this package (evidence already imports
// workflow — the reverse edge would be a cycle).
const FillMarker = orchestration.FillMarker

// FillSlot renders a substance slot, e.g. FillSlot("the impl file path") ->
// "<FILL: the impl file path>".
func FillSlot(desc string) string { return orchestration.FillSlot(desc) }
