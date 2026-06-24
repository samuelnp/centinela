package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gitutil"
)

// TestRenderDeliveryChoiceWithOptions lists each offered option's command.
func TestRenderDeliveryChoiceWithOptions(t *testing.T) {
	out := RenderDeliveryChoice("alpha", []gitutil.Option{gitutil.OptionPR, gitutil.OptionMerge})
	for _, want := range []string{"alpha", "deliver alpha --via pr", "deliver alpha --via merge"} {
		if !strings.Contains(out, want) {
			t.Fatalf("panel missing %q:\n%s", want, out)
		}
	}
}

// TestRenderDeliveryChoiceEmpty notes there is no delivery target.
func TestRenderDeliveryChoiceEmpty(t *testing.T) {
	out := RenderDeliveryChoice("beta", nil)
	if !strings.Contains(strings.ToLower(out), "no delivery target") {
		t.Fatalf("empty panel should note no delivery target:\n%s", out)
	}
	if strings.Contains(out, "--via") {
		t.Fatalf("empty panel should list no options:\n%s", out)
	}
}
