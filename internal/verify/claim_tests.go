package verify

import (
	"fmt"
	"strings"
	"time"

	"github.com/samuelnp/centinela/internal/acceptance"
	"github.com/samuelnp/centinela/internal/config"
)

const claimTestsPass = "tests-pass"

// priorRunLabel names the reused single-run outcome. It is a LABEL, not a
// command string — nothing may classify it as if it were one.
const priorRunLabel = "validate.commands"

// checkTestsPass re-runs the configured validate commands and confirms they
// exit 0 AND — for acceptance-classified commands — that their report shows
// scenarios actually ran. Missing/misconfigured commands surface as
// CONFIG-ERROR; a timeout as TIMEOUT; a non-zero exit as FAIL with the failing
// command named; detected acceptance skips as FAIL naming the counts — each
// distinct so the operator can tell a fabricated claim from a broken
// environment.
func checkTestsPass(cfg *config.Config, deps Deps, role string, timeout time.Duration) Check {
	c := Check{Claim: claimTestsPass, Role: role}
	if len(cfg.Validate.Commands) == 0 {
		c.Status = StatusConfigError
		c.Detail = "no validate.commands configured — set validate.commands in centinela.toml"
		return c
	}
	policy := cfg.Validate.AcceptanceSkipPolicy
	if deps.PriorTestRun != nil {
		// The reused outcome is labelled "validate.commands", not a real command
		// string, so per-command classification cannot apply to it. Classify over
		// the configured set as a whole — getting this wrong silently disarms the
		// rule on the exact path `complete` uses.
		scope := acceptance.ScopeNone
		if acceptance.AnyExecutionCommand(cfg.Validate.Commands) {
			scope = reusedScope(cfg.Validate.Commands)
		}
		return classifyTestRun(c, priorRunLabel, *deps.PriorTestRun, scope, policy).check
	}
	return runEachCommand(c, cfg, deps, policy, timeout)
}

// runEachCommand returns the first non-PASS classification, or an aggregate
// PASS carrying every advisory note the passing commands produced, so a
// limitation is never swallowed by the summary line.
func runEachCommand(c Check, cfg *config.Config, deps Deps, policy string, timeout time.Duration) Check {
	var notes []string
	for _, cmd := range cfg.Validate.Commands {
		out := deps.Runner.Run(deps.Root, cmd, timeout)
		v := classifyTestRun(c, cmd, out, acceptance.ScopeOf(cmd), policy)
		if v.check.Status != StatusPass {
			return v.check
		}
		if v.note != "" {
			notes = append(notes, v.note)
		}
	}
	c.Status = StatusPass
	c.Detail = "all validate.commands exited 0"
	if len(notes) > 0 {
		c.Detail += " — " + strings.Join(notes, "; ")
	}
	return c
}

// classifyTestRun maps one command outcome onto a Check, keeping config errors,
// timeouts and claim failures distinct. An exit-0 acceptance run is additionally
// judged on its REPORT, so `centinela verify` cannot certify a tests-pass claim
// the validate path would now reject. policy is passed in explicitly: the
// classifier never reads config itself.
func classifyTestRun(c Check, cmd string, out RunOutcome, scope acceptance.Scope, policy string) runVerdict {
	switch {
	case out.TimedOut:
		c.Status = StatusTimeout
		c.Detail = fmt.Sprintf("command %q exceeded the verify timeout", cmd)
	case out.StartErr != nil:
		c.Status = StatusConfigError
		c.Detail = fmt.Sprintf("command %q could not start: %s", cmd, missingBinary(out.StartErr))
	case out.ExitCode != 0:
		c.Status = StatusFail
		c.Detail = fmt.Sprintf("command %q exited %d", cmd, out.ExitCode)
	default:
		return classifyAcceptanceRun(c, cmd, out, scope, policy)
	}
	return runVerdict{check: c}
}

// missingBinary extracts a human-readable cause from a command start error.
func missingBinary(err error) string {
	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		return msg
	}
	return msg
}
