package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

// ValidateArtifacts checks that the required artifacts exist before completing a step.
func ValidateArtifacts(feature, step string, cfg *config.Config) error {
	switch step {
	case "plan":
		return validatePlan(feature)
	case "tests":
		return validateTests(feature, cfg)
	case "validate":
		if err := validateGatekeeper(feature); err != nil {
			return err
		}
		return validateProductionReadiness(feature, cfg)
	}
	return nil
}

func validatePlan(feature string) error {
	planFile := fmt.Sprintf("docs/plans/%s.md", feature)
	if _, err := os.Stat(planFile); err != nil {
		return fmt.Errorf("plan file not found: %s", planFile)
	}
	specs, _ := filepath.Glob("specs/*.feature")
	if len(specs) == 0 {
		return fmt.Errorf("no .feature spec found in specs/")
	}
	return nil
}

func validateProductionReadiness(feature string, cfg *config.Config) error {
	if !cfg.Gates.ProductionReadinessEnabled {
		return nil
	}
	report := fmt.Sprintf(".workflow/%s-production-readiness.md", feature)
	data, err := os.ReadFile(report)
	if err != nil {
		return fmt.Errorf("production readiness report not found: %s\nRun the subagent first", report)
	}
	return checkPRStatus(string(data), feature)
}

func checkPRStatus(content, feature string) error {
	if strings.Contains(content, "**Status:** BLOCKING") {
		return fmt.Errorf(
			"production readiness: BLOCKING\nFix CRITICAL issues in %q, then re-run the subagent.\nOr: centinela start %s-hardening",
			feature, feature,
		)
	}
	return nil
}

// ProductionReadinessWarning returns feature if report status is WARNING, else "".
func ProductionReadinessWarning(feature string, cfg *config.Config) string {
	if !cfg.Gates.ProductionReadinessEnabled {
		return ""
	}
	report := fmt.Sprintf(".workflow/%s-production-readiness.md", feature)
	data, err := os.ReadFile(report)
	if err != nil {
		return ""
	}
	if strings.Contains(string(data), "**Status:** WARNING") {
		return feature
	}
	return ""
}

func validateGatekeeper(feature string) error {
	report := fmt.Sprintf(".workflow/%s-gatekeeper.md", feature)
	if _, err := os.Stat(report); err != nil {
		return fmt.Errorf("gatekeeper report not found: %s", report)
	}
	return nil
}
