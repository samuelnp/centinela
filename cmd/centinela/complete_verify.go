package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/verify"
)

// runClaimVerification re-derives ground truth for the step's evidence claims
// and hard-blocks completion on any failing claim. Warnings (e.g. the heuristic
// edge-case-to-test mapping) are surfaced but do not block. On a hard block it
// records a verify-rejection telemetry event before returning the error.
//
// prior, when non-nil, is the machine record of a suite run this same process
// just completed (the validate gate's own executeValidation); it suppresses
// the tests-pass re-run. Every other surface passes nil and re-runs for real.
func runClaimVerification(feature, step, model string, cfg *config.Config, prior *verify.RunOutcome) error {
	// Contract-aware, like the gate and the directive: a legacy workflow's
	// validate claims live under validation-specialist, not gatekeeper.
	deps := verifyDepsFor(feature, step)
	deps.PriorTestRun = prior
	res := verify.Verify(feature, step, cfg, deps)
	fmt.Println(ui.RenderVerification(res))
	if res.HasFailures() {
		emitVerifyRejection(cfg, feature, step, res, model)
		return fmt.Errorf("claim verification failed for %q — evidence diverges from ground truth", feature)
	}
	return nil
}
