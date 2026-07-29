// Acceptance: specs/adversarial-validate-verifier.feature
//
// Report-writing helpers shared by the scenario files. Split out of
// adversarial_validate_verifier_helper_test.go to keep both under G1.
package acceptance_test

import (
	"fmt"
	"path/filepath"
	"testing"
)

// avvSeedWorkflow writes a minimal workflow state at the validate step.
// orchestrationMode is deliberately left empty (non-strict) so evidence JSON
// is never required — only the gatekeeper .md report gates this step.
func avvSeedWorkflow(t *testing.T, dir, feature, contract string) {
	t.Helper()
	body := fmt.Sprintf(`{"feature":%q,"currentStep":"validate",`+
		`"stepOrder":["plan","code","tests","validate","docs"],"steps":{},`+
		`"validateContract":%q}`, feature, contract)
	mustWrite(t, filepath.Join(dir, ".workflow", feature+".json"), body)
}

// avvWriteReport writes the gatekeeper report body verbatim.
func avvWriteReport(t *testing.T, dir, feature, body string) {
	t.Helper()
	mustWrite(t, filepath.Join(dir, ".workflow", feature+"-gatekeeper.md"), body)
}

// avvGroundedCommands is a commands array whose sole entry satisfies Assess's
// "one passing `centinela validate`" requirement.
const avvGroundedCommands = `[{"argv":["centinela","validate"],"exitCode":0,"durationMs":10}]`

// avvReport renders a gatekeeper report with the given status, findings
// bullets, and verification commands array (revision/treeDigest left blank
// for a subsequent `artifact stamp` to fill).
func avvReport(status, findings, commands string) string {
	return fmt.Sprintf("### Adversarial Verifier Report: demo\n**Status:** %s\n\n#### Findings\n%s\n\n"+
		"```json centinela:verification\n{\"revision\":\"\",\"treeDigest\":\"\",\"commands\":%s}\n```\n",
		status, findings, commands)
}
