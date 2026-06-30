package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
)

// Unit: a hook directive panel renders without any rounded border characters
// while keeping its content.
func TestPanelDirectiveHasNoBorder(t *testing.T) {
	out := ui.RenderBlocked("code", "plan", "my-feature", "/tmp/a.go")
	if strings.ContainsAny(out, "╭╮╰╯│") {
		t.Fatalf("blocked-write directive should have no border box, got:\n%s", out)
	}
	for _, want := range []string{"🛡️👁️", "BLOCKED WRITE", "Next action"} {
		if !strings.Contains(out, want) {
			t.Errorf("directive lost content %q", want)
		}
	}
}
