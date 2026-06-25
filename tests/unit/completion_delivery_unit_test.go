package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitutil"
)

// TestDeliveryMatrixAndDirective verifies the option matrix and that the
// directive lists exactly the offered options as deliver commands.
func TestDeliveryMatrixAndDirective(t *testing.T) {
	both := gitutil.DeliveryOptions(true, true)
	if len(both) != 2 || !gitutil.Supports(both, gitutil.OptionPR) || !gitutil.Supports(both, gitutil.OptionMerge) {
		t.Fatalf("origin+worktree should offer both, got %v", both)
	}
	if got := gitutil.DeliveryOptions(false, false); len(got) != 0 {
		t.Fatalf("neither should offer nothing, got %v", got)
	}

	d := gitutil.DeliveryDirective("alpha", gitutil.DeliveryOptions(true, false))
	if !strings.Contains(d, "centinela deliver alpha --via pr") || strings.Contains(d, "--via merge") {
		t.Fatalf("origin-only directive should offer pr only:\n%s", d)
	}

	empty := gitutil.DeliveryDirective("beta", nil)
	if !strings.Contains(empty, "no delivery target") {
		t.Fatalf("empty directive should state no target:\n%s", empty)
	}
}
