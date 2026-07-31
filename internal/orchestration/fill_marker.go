package orchestration

import (
	"fmt"
	"strings"
)

// FillMarker is the canonical substance-slot template. Used in MARKDOWN bodies
// only (companions, artifact stubs) — NEVER in an evidence JSON list field, where
// a non-real-file string would fail the actionable-outputs validator.
//
// It lives in this package rather than next to the renderer in
// internal/evidence because two sides need it: the package that RENDERS a slot
// (evidence) and the gate that must REJECT an unfilled one (workflow).
// evidence already imports workflow, so a workflow -> evidence edge would be a
// cycle; both already depend on orchestration, which imports nothing internal.
const FillMarker = "<FILL: %s>"

// fillPrefix is the invariant head of a rendered slot — the part that survives
// any description, and therefore the only part a detector may match on.
const fillPrefix = "<FILL:"

// FillSlot renders a substance slot, e.g. FillSlot("the impl file path") ->
// "<FILL: the impl file path>".
func FillSlot(desc string) string { return fmt.Sprintf(FillMarker, desc) }

// HasFillMarker reports whether s still carries an unreplaced substance slot,
// i.e. whether the text is a scaffold rather than something an author wrote.
func HasFillMarker(s string) bool { return strings.Contains(s, fillPrefix) }
