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

// UnreplacedSlot reports whether s still carries a slot an author was meant to
// replace — the one rule that separates a scaffold from something written.
//
// The question is asked of CONTENT, never of position. A caller scanning a
// document must be able to ask it of every line: a markdown heading or a
// single line of preamble is otherwise all it takes to park an untouched
// template behind a gate that only looked at line 1.
//
// A slot whose description is nothing but an ellipsis is a CITATION, not a
// slot: "<FILL: ...>" names no substance to fill in, and it is the generic
// form the CLI's own errors and docs use when they talk ABOUT the marker.
// Reading it as a quotation is what lets a changelog or report legitimately
// discuss templating without tripping the gate that enforces it. The
// alternative — judging the substance of the surrounding prose — would have to
// accept a half-filled "- fix: <FILL: one-line summary>", which this rule
// correctly refuses.
//
// An unterminated "<FILL:" is prose, not a rendered slot: FillSlot always
// emits the closing bracket.
func UnreplacedSlot(s string) bool {
	for rest := s; ; {
		i := strings.Index(rest, fillPrefix)
		if i < 0 {
			return false
		}
		rest = rest[i+len(fillPrefix):]
		end := strings.Index(rest, ">")
		if end < 0 {
			return false
		}
		if !isSlotCitation(rest[:end]) {
			return true
		}
		rest = rest[end+1:]
	}
}

// isSlotCitation reports whether a slot description is an elision rather than
// a real instruction: empty, or nothing but ellipsis characters.
func isSlotCitation(desc string) bool {
	return strings.Trim(strings.TrimSpace(desc), ".…") == ""
}
