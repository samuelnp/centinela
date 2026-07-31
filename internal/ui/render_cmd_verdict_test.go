package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
)

func TestRenderCmdVerdict_StatusIcons(t *testing.T) {
	cases := []struct {
		status gates.Status
		icon   string
	}{
		{gates.Pass, IconDone},
		{gates.Warn, "⚠"},
		{gates.Fail, "✗"},
	}
	for _, c := range cases {
		got := RenderCmdVerdict("go test ./...", c.status, "", "")
		if !strings.Contains(got, c.icon) {
			t.Fatalf("status %v must render %q, got %q", c.status, c.icon, got)
		}
	}
}

func TestRenderCmdVerdict_DetailIsShownForEveryStatus(t *testing.T) {
	for _, s := range []gates.Status{gates.Pass, gates.Warn, gates.Fail} {
		got := RenderCmdVerdict("cmd", s, "1 skipped of 3 scenarios", "")
		if !strings.Contains(got, "1 skipped of 3 scenarios") {
			t.Fatalf("detail must be rendered for status %v, got %q", s, got)
		}
	}
}

// Output is echoed only where it is the thing to act on: a failure. A warning
// names its reason in the detail and must not dump the whole buffer.
func TestRenderCmdVerdict_OutputEchoedOnlyOnFailure(t *testing.T) {
	const output = "some very long captured output"
	if got := RenderCmdVerdict("cmd", gates.Fail, "", output); !strings.Contains(got, output) {
		t.Fatalf("a failure must echo its output, got %q", got)
	}
	for _, s := range []gates.Status{gates.Pass, gates.Warn} {
		if got := RenderCmdVerdict("cmd", s, "note", output); strings.Contains(got, output) {
			t.Fatalf("status %v must not echo the output, got %q", s, got)
		}
	}
}

// RenderCmdResult keeps its bool-shaped contract by delegating.
func TestRenderCmdResult_DelegatesToTheVerdictRenderer(t *testing.T) {
	if got := RenderCmdResult("cmd", true, "ignored"); !strings.Contains(got, IconDone) ||
		strings.Contains(got, "ignored") {
		t.Fatalf("a passing result is a bare green line, got %q", got)
	}
	if got := RenderCmdResult("cmd", false, "boom"); !strings.Contains(got, "✗") ||
		!strings.Contains(got, "boom") {
		t.Fatalf("a failing result echoes its output, got %q", got)
	}
}
