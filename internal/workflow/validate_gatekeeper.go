package workflow

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/gatereport"
)

// GatekeeperReportPath is the validate-step verifier report for feature.
func GatekeeperReportPath(feature string) string {
	return fmt.Sprintf(".workflow/%s-gatekeeper.md", feature)
}

// validateGatekeeper gates the validate step on the verifier report. Legacy
// workflows keep the historical existence-only check verbatim; workflows
// pinned to adversarial-v1 additionally require a grounded verdict. Content
// checks only — the git-backed freshness check lives in VerificationFresh so
// `hook context` never pays for a git invocation per workflow per prompt (D6).
func validateGatekeeper(feature string) error {
	report := GatekeeperReportPath(feature)
	data, err := os.ReadFile(report)
	if err != nil {
		return fmt.Errorf("gatekeeper report not found: %s", report)
	}
	if !featureUsesAdversarialVerifier(feature) {
		return nil
	}
	if err := gatereport.Assess(string(data)); err != nil {
		return fmt.Errorf("%s: %w", report, err)
	}
	return nil
}
