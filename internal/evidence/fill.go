package evidence

import "fmt"

// FillMarker is the canonical substance-slot template. Used in MARKDOWN bodies
// only (companions, artifact stubs) — NEVER in an evidence JSON list field, where
// a non-real-file string would fail the actionable-outputs validator.
const FillMarker = "<FILL: %s>"

// FillSlot renders a substance slot, e.g. FillSlot("the impl file path") ->
// "<FILL: the impl file path>".
func FillSlot(desc string) string { return fmt.Sprintf(FillMarker, desc) }
