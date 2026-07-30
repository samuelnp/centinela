package acceptance

import "fmt"

// Skip-policy values for [validate] acceptance_skip_policy. The same three
// literals are normalized and validated in internal/config; they are restated
// here rather than imported so this package stays a stdlib-only leaf.
// cmd/centinela pins the two sets equal.
const (
	PolicyFail = "fail"
	PolicyWarn = "warn"
	PolicyOff  = "off"
)

// Verdict is the acceptance-skip judgement for one command, in this leaf's own
// vocabulary so the package needs no gate or UI import.
type Verdict int

const (
	VerdictPass Verdict = iota
	VerdictWarn
	VerdictFail
)

// UnparseableDetail is the honest reading of output that matches no supported
// runner shape: an undetected skip is not a proven skip, so it is never a
// failure — and a suite that asserted nothing is not a pass, so it is never
// silently green either.
const UnparseableDetail = "acceptance report could not be parsed — skips undetected; " +
	"run with `go test -json` / `-v` or a cucumber-compatible summary"

// Judge applies the skip policy to ONE acceptance-classified command that
// already exited 0. Callers must handle a non-zero exit before calling: an exit
// failure is reported as the exit failure, never re-labelled as a skip verdict.
func Judge(cmd, output, policy string) (Verdict, string) {
	if policy == PolicyOff {
		return VerdictPass, ""
	}
	s, ok := Detect(output)
	if !ok {
		return VerdictWarn, UnparseableDetail
	}
	if !s.SkipData {
		return VerdictPass, fmt.Sprintf(
			"%s carries no skip data — add -json or -v to make skips detectable", s.Shape)
	}
	if s.Unexecuted() == 0 && s.Scenarios > 0 {
		return VerdictPass, ""
	}
	if policy == PolicyWarn {
		return VerdictWarn, Describe(cmd, s)
	}
	return VerdictFail, Describe(cmd, s)
}

// Describe names the counts, the shape and the command, so the operator can act
// on the message without re-running anything.
func Describe(cmd string, s Summary) string {
	if s.Scenarios == 0 {
		return fmt.Sprintf("%q executed no scenarios at all (%s)", cmd, s.Shape)
	}
	return fmt.Sprintf("%q reported %d skipped, %d pending, %d undefined of %d scenarios (%s)",
		cmd, s.Skipped, s.Pending, s.Undefined, s.Scenarios, s.Shape)
}
