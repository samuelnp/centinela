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
func runClaimVerification(feature, step, model string, cfg *config.Config) error {
	res := verify.Verify(feature, step, cfg, verify.Deps{
		Root:   verifyRoot(),
		Runner: verify.NewExecRunner(),
	})
	fmt.Println(ui.RenderVerification(res))
	if res.HasFailures() {
		emitVerifyRejection(cfg, feature, step, res, model)
		return fmt.Errorf("claim verification failed for %q — evidence diverges from ground truth", feature)
	}
	return nil
}
