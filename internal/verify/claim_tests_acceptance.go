package verify

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/acceptance"
)

// runVerdict is one command's classification plus the advisory note a PASS
// carried (a parse limitation, or skip counts the policy chose not to fail on).
// The note is a FIELD, not something a caller re-parses out of Detail.
type runVerdict struct {
	check Check
	note  string
}

// classifyAcceptanceRun judges an exit-0 run. A non-acceptance command is a
// plain pass: classification gates the parse and the parse gates the verdict,
// so a legitimately skipping unit or integration tier can never be failed here.
//
// An UNPARSEABLE report stays StatusPass with the limitation named in the note.
// The verifier must not invent a failure it cannot prove — an undetected skip
// is not a proven skip.
func classifyAcceptanceRun(c Check, cmd string, out RunOutcome, scope acceptance.Scope, policy string) runVerdict {
	c.Status = StatusPass
	c.Detail = fmt.Sprintf("command %q exited 0", cmd)
	if out.AcceptanceJudged != nil {
		return inheritJudgement(c, cmd, *out.AcceptanceJudged)
	}
	if scope == acceptance.ScopeNone {
		return runVerdict{check: c}
	}
	verdict, detail := acceptance.JudgeScoped(cmd, out.Output, policy, scope)
	if verdict == acceptance.VerdictFail {
		c.Status = StatusFail
		c.Detail = fmt.Sprintf("acceptance skips detected — %s", detail)
		return runVerdict{check: c}
	}
	if detail != "" {
		c.Detail = fmt.Sprintf("command %q exited 0 — %s", cmd, detail)
	}
	return runVerdict{check: c, note: detail}
}

// reusedScope resolves the scope for the joined transcript behind priorRunLabel,
// which is not a command and so has no ScopeOf answer of its own. A single
// whole-repo command anywhere in the set makes the whole transcript mixed —
// the reading that cannot over-block another tier's legitimate skips. When every
// acceptance command is itself acceptance-scoped, nothing needs filtering and a
// runner report with no go-package blocks (a bare cucumber run) still counts.
func reusedScope(commands []string) acceptance.Scope {
	scope := acceptance.ScopeNone
	for _, cmd := range commands {
		switch acceptance.ScopeOf(cmd) {
		case acceptance.ScopeMixed:
			return acceptance.ScopeMixed
		case acceptance.ScopeAcceptance:
			scope = acceptance.ScopeAcceptance
		}
	}
	return scope
}

// inheritJudgement reuses the producer's per-command analysis instead of
// re-deriving a weaker answer. The producer judged each command with THAT
// command's own scope — information a joined transcript no longer carries — so
// re-parsing here could only be less correct. When no analysis ran, the detail
// says exactly that rather than blaming an unparseable report.
func inheritJudgement(c Check, cmd string, j AcceptanceJudgement) runVerdict {
	if j.Failed {
		c.Status = StatusFail
		c.Detail = fmt.Sprintf("acceptance skips detected — %s", j.Detail)
		return runVerdict{check: c}
	}
	note := "acceptance skip analysis was not performed — no acceptance-classified command"
	switch {
	case j.Analysed:
		note = "acceptance skip analysis already performed per command by the validate gate: " + j.Detail
	case j.Detail != "":
		// Not analysed, but the producer said WHY (e.g. the off policy).
		note = "acceptance skip analysis was not performed — " + j.Detail
	}
	c.Status = StatusPass
	c.Detail = fmt.Sprintf("command %q exited 0 — %s", cmd, note)
	return runVerdict{check: c, note: note}
}
