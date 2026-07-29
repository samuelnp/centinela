package gatereport

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateArgv is the command a grounded verdict MUST record as passing.
const ValidateArgv = "centinela validate"

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

func assessGrounding(report string) error {
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

// hasPassingValidate accepts any PATH to a centinela binary, because a
// verifier working in a worktree must build its own (the installed binary
// lags the branch and lacks new subcommands). Requiring the literal
// "centinela validate" would pressure an honest verifier into recording an
// argv it did not run. The basename must still start with "centinela", so an
// arbitrarily named scratch binary does not satisfy the check.
func hasPassingValidate(commands []Command) bool {
	for _, c := range commands {
		if len(c.Argv) != 2 || c.Argv[1] != "validate" || c.ExitCode != 0 {
			continue
		}
		if strings.HasPrefix(filepath.Base(c.Argv[0]), "centinela") {
			return true
		}
	}
	return false
}
