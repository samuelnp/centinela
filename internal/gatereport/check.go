package gatereport

import (
	"fmt"
	"strings"
)

// ValidateArgv is the command a grounded verdict MUST record as passing.
const ValidateArgv = "centinela validate"

const remedy = "Re-run the verifier (see docs/architecture/gatekeeper-prompt.md)."

// Assess reports why a report is inadmissible as a validate-step verdict, or
// nil when it is grounded. Every failure path blocks: a missing or unparseable
// verdict is never treated as a pass.
func Assess(report string) error {
	switch Normalize(Verdict(report)) {
	case "":
		return fmt.Errorf("gatekeeper verdict is missing or unparseable — expected a `**Status:** SAFE | WARNING | CRITICAL` line. %s", remedy)
	case VerdictCritical:
		return fmt.Errorf("gatekeeper verdict: CRITICAL — %s. Re-verify with a FRESH verifier context after fixing", FirstFinding(report))
	}
	v, err := ParseVerification(report)
	if err != nil {
		return fmt.Errorf("gatekeeper report has no commands-run record — a narrated verdict is not evidence. %s", remedy)
	}
	return assessRecord(v)
}

// assessRecord enforces the grounding contract on a parsed record.
func assessRecord(v Verification) error {
	if len(v.Commands) == 0 {
		return fmt.Errorf("gatekeeper report has no commands-run record — a narrated verdict is not evidence. %s", remedy)
	}
	lines := make([]string, 0, len(v.Commands))
	for _, c := range v.Commands {
		if len(c.Argv) == 0 {
			return fmt.Errorf("gatekeeper report has a commands-run record entry with an empty argv. %s", remedy)
		}
		lines = append(lines, c.Line())
	}
	if !hasPassingValidate(v.Commands) {
		return fmt.Errorf("gatekeeper report has no commands-run record of a passing %q (recorded: %s). %s",
			ValidateArgv, strings.Join(lines, "; "), remedy)
	}
	if v.Revision == "" || v.TreeDigest == "" {
		return fmt.Errorf("gatekeeper report records no verified revision/treeDigest — run `centinela artifact stamp` as the verifier's LAST action. %s", remedy)
	}
	return nil
}

func hasPassingValidate(commands []Command) bool {
	for _, c := range commands {
		if c.Line() == ValidateArgv && c.ExitCode == 0 {
			return true
		}
	}
	return false
}
