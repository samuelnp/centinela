package main

import "github.com/samuelnp/centinela/internal/verify"

// completedValidationOutcome synthesizes the machine record of the validate
// gate's own executeValidation() run, for reuse by claim verification's
// tests-pass check in the SAME complete invocation.
//
// Soundness invariant: this is only reachable AFTER executeValidation()
// returned nil in this very process — every validate.commands entry (the full
// profiled suite included) exited 0 seconds earlier — and after a second
// workflow.VerificationFresh check confirmed the tree did not change while
// the suite ran. No on-disk, agent-writable state is involved; the outcome
// exists only in this process's memory.
//
// It carries the REAL captured transcript and the acceptance-skip judgement
// that run already produced, so the reuse path neither re-parses a lossy
// concatenation nor claims an analysis it did not perform. A fixed prose
// Output would have made the skip rule structurally unable to fire here.
func completedValidationOutcome() *verify.RunOutcome {
	return &verify.RunOutcome{
		ExitCode:         0,
		Output:           lastValidateRun.describeCommands(completedValidationHeader),
		AcceptanceJudged: lastValidateRun.judgement(),
	}
}

// completedValidationHeader names the provenance of the reused record; the
// commands' own output follows it.
const completedValidationHeader = "validate.commands executed by this complete gate at the verified tree"
