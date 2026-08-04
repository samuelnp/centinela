package ui

import (
	"strings"
	"testing"
)

// trimCase pins one input/output pair for trimTrailingWS. Cases marked latent
// document behaviour no in-tree caller can reach today; they exist so a future
// change that DOES reach them shows up as a visible diff in this table rather
// than as a silent rendering change.
type trimCase struct {
	name   string
	in     string
	want   string
	latent string
}

var trimCases = []trimCase{
	{name: "empty string", in: "", want: ""},
	{name: "spaces only collapse to an empty line", in: "   ", want: ""},
	{name: "tabs only collapse to an empty line", in: "\t\t", want: ""},
	{name: "single line, no trailing newline", in: "abc  ", want: "abc"},
	{name: "already trimmed is unchanged", in: "abc", want: "abc"},
	{name: "mixed spaces and tabs", in: "abc \t \t", want: "abc"},
	// A caller appending a panel to a stream relies on the final newline
	// surviving; TrimRight is applied per line, never to the line count.
	{name: "one trailing newline survives", in: "abc\n", want: "abc\n"},
	{name: "several trailing newlines survive", in: "abc\n\n\n", want: "abc\n\n\n"},
	// The exact shape lipgloss.JoinVertical produces for a spacer "" between
	// padded siblings: the blank separator must stay a line, just an empty one.
	{name: "interior blank separator stays a line", in: "a  \n   \nb  ", want: "a\n\nb"},
	{
		name: "CRLF line endings are left untouched", in: "a  \r\nb  \r\n", want: "a  \r\nb  \r\n",
		latent: "trimTrailingWS splits on \\n only and TrimRight stops at \\r, so the " +
			"spaces before each \\r survive. Unreachable today: no render function in " +
			"internal/ui or cmd/centinela/hook_context* emits a literal \\r.",
	},
	{
		name: "a meaningful trailing tab is eaten", in: "col1\tcol2\t", want: "col1\tcol2",
		latent: "tabs are in the cutset, so a tab used as content (not padding) would be " +
			"stripped. Unreachable today: no panel renders tab-separated columns.",
	},
	{
		name: "non-breaking space is preserved", in: "abc\u00a0\u00a0", want: "abc\u00a0\u00a0",
		latent: "U+00A0 is not in the ASCII cutset, so NBSP survives as content. " +
			"Unreachable today: nothing in the touched render path emits NBSP.",
	},
}

func TestTrimTrailingWSTable(t *testing.T) {
	for _, tc := range trimCases {
		t.Run(tc.name, func(t *testing.T) {
			got := trimTrailingWS(tc.in)
			if got != tc.want {
				t.Fatalf("trimTrailingWS(%q) = %q, want %q%s", tc.in, got, tc.want, latentNote(tc))
			}
		})
	}
}

// TestTrimTrailingWSLongLine keeps the 5000-byte case out of the table so a
// failure prints a length, not five thousand characters.
func TestTrimTrailingWSLongLine(t *testing.T) {
	body := strings.Repeat("x", 5000)
	got := trimTrailingWS(body + "   \t")
	if got != body {
		t.Fatalf("long line mangled: got %d bytes, want %d", len(got), len(body))
	}
}

// TestTrimTrailingWSIdempotent guards the property the six render functions
// depend on: wrapping an already-wrapped return must be a no-op, so a double
// wrap (or a re-render of stored output) can never shift bytes.
func TestTrimTrailingWSIdempotent(t *testing.T) {
	inputs := []string{"", "   ", "abc  ", "a  \n   \nb  ", "abc\n", "a  \r\nb  \r\n", panelDietRender()}
	for _, tc := range trimCases {
		inputs = append(inputs, tc.in)
	}
	for _, in := range inputs {
		once := trimTrailingWS(in)
		if twice := trimTrailingWS(once); twice != once {
			t.Fatalf("trimTrailingWS not idempotent for %q: once=%q twice=%q", in, once, twice)
		}
	}
}

func latentNote(tc trimCase) string {
	if tc.latent == "" {
		return ""
	}
	return "\n(pinned current behaviour: " + tc.latent + ")"
}
