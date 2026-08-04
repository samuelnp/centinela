package gatereport

import "fmt"

const remedy = "Re-run the verifier (see docs/architecture/gatekeeper-prompt.md)."

// Assess reports why a report is inadmissible as a validate-step verdict, or
// nil when it is grounded. Every failure path blocks: a missing or unparseable
// verdict is never treated as a pass.
//
// A CRITICAL verdict reports the grounding failure ALONGSIDE the finding, so a
// fail-closed scaffolded stub (which ships CRITICAL and an empty commands
// array) still tells the operator it has no commands-run record.
func Assess(report string) error {
	verdict := Normalize(Verdict(report))
	if verdict == "" {
		return fmt.Errorf("gatekeeper verdict is missing or unparseable — expected a `**Status:** SAFE | WARNING | CRITICAL` line. %s", remedy)
	}
	grounding := assessGrounding(report)
	if verdict == VerdictCritical {
		return criticalError(report, grounding)
	}
	return grounding
}

func criticalError(report string, grounding error) error {
	detail := ""
	if grounding != nil {
		detail = fmt.Sprintf(" (and %s)", grounding)
	}
	return fmt.Errorf("gatekeeper verdict: CRITICAL — %s%s. Re-verify with a FRESH verifier context after fixing",
		FirstFinding(report), detail)
}
