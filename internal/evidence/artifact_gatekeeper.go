package evidence

import "fmt"

// gatekeeperBody renders the validate-step gatekeeper report stub. The
// `**Status:**` line is parsed by `centinela validate` so the literal
// SAFE | WARNING | BLOCKING vocabulary is preserved verbatim. "Analyzed Specs"
// is mechanically pre-filled from specs/*.feature (see analyzedSpecsList).
func gatekeeperBody(feature string) []byte {
	return []byte(fmt.Sprintf(`### Gatekeeper Report: %s
**Date:** %s
**Status:** SAFE

#### Analyzed Specs
%s

#### Findings
- %s

#### Recommendation
- %s
`, feature, today(), analyzedSpecsList(), FillSlot("affected spec / scenario / risk / suggestion; empty if SAFE"), FillSlot("SAFE/WARNING/BLOCKING + one-line rationale")))
}
