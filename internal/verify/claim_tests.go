package verify

import (
	"fmt"
	"strings"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

const claimTestsPass = "tests-pass"

// checkTestsPass re-runs the configured validate commands and confirms they
// exit 0. Missing/misconfigured commands surface as CONFIG-ERROR; a timeout as
// TIMEOUT; a non-zero exit as FAIL with the failing command named — each
// distinct from the others so the operator can tell a fabricated claim from a
// broken environment.
func checkTestsPass(cfg *config.Config, deps Deps, role string, timeout time.Duration) Check {
	c := Check{Claim: claimTestsPass, Role: role}
	if len(cfg.Validate.Commands) == 0 {
		c.Status = StatusConfigError
		c.Detail = "no validate.commands configured — set validate.commands in centinela.toml"
		return c
	}
	if deps.PriorTestRun != nil {
		return classifyTestRun(c, "validate.commands", *deps.PriorTestRun)
	}
	for _, cmd := range cfg.Validate.Commands {
		out := deps.Runner.Run(deps.Root, cmd, timeout)
		if res := classifyTestRun(c, cmd, out); res.Status != StatusPass {
			return res
		}
	}
	c.Status = StatusPass
	c.Detail = "all validate.commands exited 0"
	return c
}

// classifyTestRun maps one command outcome onto a Check, keeping config errors,
// timeouts and claim failures distinct.
func classifyTestRun(c Check, cmd string, out RunOutcome) Check {
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
		c.Status = StatusPass
		c.Detail = fmt.Sprintf("command %q exited 0", cmd)
	}
	return c
}

// missingBinary extracts a human-readable cause from a command start error.
func missingBinary(err error) string {
	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		return msg
	}
	return msg
}
