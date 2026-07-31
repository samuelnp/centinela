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
	return JudgeScoped(cmd, output, policy, ScopeOf(cmd))
}

// JudgeScoped is Judge with the scope supplied explicitly. Use it when the
// command string is a LABEL rather than a real command (the reused
// "validate.commands" outcome), where ScopeOf cannot answer and the conservative
// ScopeMixed is the only honest reading of a joined transcript.
func JudgeScoped(cmd, output, policy string, scope Scope) (Verdict, string) {
	if policy == PolicyOff {
		return VerdictPass, ""
	}
	s, ok := Detect(output, scope)
	if !ok {
		return VerdictWarn, UnparseableDetail
	}
	if !s.SkipData {
		return VerdictPass, fmt.Sprintf(
			"%s carries no skip data — add -json or -v to make skips detectable", s.Shape)
	}
	if !failable(s, scope) {
		return VerdictPass, ""
	}
	if policy == PolicyWarn {
		return VerdictWarn, Describe(cmd, s, scope)
	}
	return VerdictFail, Describe(cmd, s, scope)
}

// failable decides whether the parsed report warrants the policy.
//
// A Gherkin summary reporting zero scenarios is failable ON ITS OWN. It must
// not be rescued by a co-occurring passing signal: a Go wrapper test's
// `--- PASS:` proves the wrapper ran, never that a scenario did, and that
// combination is the ordinary shape of godog driven from `go test`.
//
// The merged-total "nothing executed" rule stays limited to an acceptance-scoped
// command. On a whole-repo command, zero attributed results means the acceptance
// tier was not identifiable in that output — not that nothing ran — and failing
// on it would be the same over-block Scope exists to prevent.
func failable(s Summary, scope Scope) bool {
	if s.Unexecuted() > 0 || s.GherkinZero {
		return true
	}
	return scope == ScopeAcceptance && s.Scenarios == 0
}

// Describe names the counts, the shape and the command, so the operator can act
// on the message without re-running anything. The attribution clause is gated
// on Summary.Attributed — set by the parsers that actually filtered — so the
// message can never assert an attribution that did not happen.
func Describe(cmd string, s Summary, _ Scope) string {
	suffix := ""
	if s.Attributed {
		suffix = ", counts attributed to " + AcceptancePath + " only"
	}
	switch {
	case s.GherkinZero:
		return fmt.Sprintf("%q reported 0 scenarios — no scenarios were executed (%s%s)",
			cmd, s.Shape, suffix)
	case s.Scenarios == 0:
		return fmt.Sprintf("%q executed no scenarios at all (%s%s)", cmd, s.Shape, suffix)
	}
	return fmt.Sprintf("%q reported %d skipped, %d pending, %d undefined of %d scenarios (%s%s)",
		cmd, s.Skipped, s.Pending, s.Undefined, s.Scenarios, s.Shape, suffix)
}
