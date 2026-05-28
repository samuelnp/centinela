package evidence

import "fmt"

// edgeCasesBody renders the .workflow/<feature>-edge-cases.md stub the QA
// senior fleshes out at the tests step.
func edgeCasesBody(feature string) []byte {
	return []byte(fmt.Sprintf(`# Edge Cases: %s

## Covered

- _List each edge case the test suite exercises._

## Residual Risks

- _List anything intentionally out of scope, with mitigation notes._
`, feature))
}
