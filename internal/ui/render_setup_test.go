package ui

import (
	"strings"
	"testing"
)

// TestRenderBrownfieldSetupNeeded asserts the brownfield panel carries the
// enrich-then-confirm directive literals the hook and spec depend on.
func TestRenderBrownfieldSetupNeeded(t *testing.T) {
	out := RenderBrownfieldSetupNeeded()
	for _, want := range []string{
		"BROWNFIELD",                  // panel title: BROWNFIELD PROJECT DETECTED
		"Do NOT interrogate",          // do not cold-question the user
		"analyze",                     // centinela analyze step
		"synthesize",                  // centinela synthesize step
		"ENRICH",                      // enrich the draft from source
		"**Project Stage:** existing", // stamp the existing stage
		"confirm",                     // confirm uncertain fields before finalize
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderBrownfieldSetupNeeded missing %q in:\n%s", want, out)
		}
	}
}
