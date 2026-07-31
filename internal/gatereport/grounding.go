package gatereport

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateArgv is the command a grounded verdict MUST record as passing.
const ValidateArgv = "centinela validate"

func assessGrounding(report string) error {
	v, err := ParseVerification(report)
	if err != nil {
		return groundingParseError(err)
	}
	return assessRecord(v)
}

// groundingParseError keeps the operator-facing "no commands-run record"
// remedy for every unparseable block, but names WHAT is malformed when a block
// is present and only its shape is wrong. "No record" on its own would send a
// verifier hunting for a fence it can plainly see.
func groundingParseError(err error) error {
	if errors.Is(err, ErrNoBlock) {
		return fmt.Errorf("gatekeeper report has no commands-run record — a narrated verdict is not evidence. %s", remedy)
	}
	return fmt.Errorf("gatekeeper report has no commands-run record it can trust — %s. %s", err, remedy)
}

// assessRecord enforces the ADMISSIBILITY contract on a parsed record — a
// different question from ValidateCommandsSchema's shape rule, which
// ParseVerification has already applied upstream. The empty-argv loop below
// is kept as cheap defense-in-depth for any future caller that builds a
// Verification without going through the parser; it is not a second copy of
// the shape rule.
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
