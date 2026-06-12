package evidence

import "fmt"

// prodReadyBody renders the production-readiness review stub. The
// `**Status:**` line is parsed by `centinela validate` so the literal
// PASS | WARNING | BLOCKING vocabulary matches the prompt template.
func prodReadyBody(feature string) []byte {
	return []byte(fmt.Sprintf(`### Production Readiness Report: %s
**Date:** %s
**Status:** PASS

#### Files Reviewed
- %s

#### Findings
| Check | File | Severity | Issue | Suggested Fix |
|-------|------|----------|-------|---------------|

#### Summary
CRITICAL: 0, WARNING: 0

#### Recommendation
PASS: confirm the feature is production-ready.
`, feature, today(), FillSlot("each file checked")))
}
