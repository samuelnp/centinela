package evidence

import "fmt"

// gatekeeperBody renders the validate-step gatekeeper report stub. The
// `**Status:**` line is parsed by `centinela validate` so the literal
// SAFE | WARNING | BLOCKING vocabulary is preserved verbatim.
func gatekeeperBody(feature string) []byte {
	return []byte(fmt.Sprintf(`### Gatekeeper Report: %s
**Date:** %s
**Status:** SAFE

#### Analyzed Specs
- _List each existing .feature file you reviewed._

#### Findings
- _Affected spec / scenario / risk / suggestion. Empty list if SAFE._

#### Recommendation
- SAFE: No conflicts detected. Proceed with implementation.
`, feature, today()))
}
