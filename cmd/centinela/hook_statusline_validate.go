package main

import (
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// validateStatusBlock classifies why the validate step cannot advance, as a
// (block, next-action) pair for the statusline. Extracted from
// hook_statusline_rules.go so neither file exceeds the size budget.
//
// Classification is by error-message substring, which is fragile: any error
// this function does not recognize falls through to MISSING_PROD_READINESS.
// Typed gate error codes are the real fix (roadmap: typed-gate-error-codes).
func validateStatusBlock(feature string, cfg *config.Config) (string, string) {
	if !fileExists(workflow.GatekeeperReportPath(feature)) {
		return "MISSING_GATEKEEPER", "run-verifier"
	}
	err := workflow.ValidateArtifacts(feature, "validate", cfg)
	if err == nil {
		return "none", "run-validate"
	}
	return classifyValidateError(err.Error())
}

// classifyValidateError maps a validate-gate failure onto a statusline block.
// Ordered most-specific first: the verifier's own failures are checked before
// the production-readiness fallback.
func classifyValidateError(msg string) (string, string) {
	switch {
	case strings.Contains(msg, "verdict: CRITICAL"):
		return "VERDICT_CRITICAL", "fix-and-reverify"
	case strings.Contains(msg, "verdict is missing or unparseable"):
		return "VERDICT_UNPARSEABLE", "rerun-verifier"
	case strings.Contains(msg, "no commands-run record"):
		return "MISSING_COMMANDS_RECORD", "rerun-verifier"
	case strings.Contains(msg, "verification is stale"):
		return "STALE_VERIFICATION", "reverify-fresh-context"
	case strings.Contains(msg, "BLOCKING"):
		return "PROD_BLOCKING", "harden-feature"
	default:
		return "MISSING_PROD_READINESS", "run-production-readiness"
	}
}
