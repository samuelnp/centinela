package evidence

import "fmt"

// edgeCasesBody renders the .workflow/<feature>-edge-cases.md stub the QA
// senior fleshes out at the tests step.
func edgeCasesBody(feature string) []byte {
	return []byte(fmt.Sprintf(`# Edge Cases: %s

## Covered

- %s

## Residual Risks

- %s
`, feature, FillSlot("each edge case the test suite exercises"), FillSlot("anything intentionally out of scope, with mitigation notes")))
}
