package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/treestate"
	"github.com/samuelnp/centinela/internal/workflow"
)

// runValidateGates runs the validate step's gates before it may advance.
// Verification is CONSTANT across every profile — NO profile branch belongs
// here; profiles scale process, never proof.
//
// Freshness runs first because it is the cheapest and the one that tells the
// operator to re-verify before they pay for a full suite run they would only
// have to repeat. Report ADMISSIBILITY (verdict + commands record) is a pure
// file read and stays in workflow.ValidateArtifacts, which wf.Complete calls.
func runValidateGates(feature, step, model string, cfg *config.Config) error {
	if err := workflow.VerificationFresh(feature, ".", treestate.NewExecRunner()); err != nil {
		telemetry.RecordCompleteRejected(cfg, feature, step, "gates", model)
		return err
	}
	if err := executeValidation(); err != nil {
		telemetry.RecordCompleteRejected(cfg, feature, step, "gates", model)
		return err
	}
	if err := runClaimVerification(feature, step, model, cfg); err != nil {
		telemetry.RecordCompleteRejected(cfg, feature, step, "verify", model)
		return err
	}
	return nil
}
