package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/verify"
)

// emitGateFailures records one gate-failure telemetry event per failing gate.
// Best-effort: Record never errors or blocks.
func emitGateFailures(cfg *config.Config, results []gates.Result, model string) {
	for _, r := range results {
		if r.Status == gates.Fail {
			telemetry.RecordGateFailure(cfg, r.Name, r.Message, model)
		}
	}
}

// emitVerifyRejection records a verify-rejection event carrying the blocking
// checks, mapping verify.Check → telemetry.CheckRef (telemetry owns its copy).
func emitVerifyRejection(cfg *config.Config, feature, step string, res verify.VerificationResult, model string) {
	telemetry.RecordVerifyRejection(cfg, feature, step, toCheckRefs(res.Failed()), model)
}

func toCheckRefs(checks []verify.Check) []telemetry.CheckRef {
	if len(checks) == 0 {
		return nil
	}
	refs := make([]telemetry.CheckRef, 0, len(checks))
	for _, c := range checks {
		refs = append(refs, telemetry.CheckRef{
			Claim:  c.Claim,
			Role:   c.Role,
			Status: string(c.Status),
			Detail: c.Detail,
		})
	}
	return refs
}
