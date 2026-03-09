package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateArtifacts checks that the required artifacts exist before completing a step.
func ValidateArtifacts(feature, step string) error {
	switch step {
	case "plan":
		return validatePlan(feature)
	case "tests":
		return validateTests()
	case "validate":
		return validateGatekeeper(feature)
	}
	return nil
}

func validatePlan(feature string) error {
	matches, _ := filepath.Glob("docs/plans/*.md")
	found := false
	for _, m := range matches {
		data, _ := os.ReadFile(m)
		if strings.Contains(string(data), feature) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no plan in docs/plans/ mentions %q", feature)
	}
	specs, _ := filepath.Glob("specs/*.feature")
	if len(specs) == 0 {
		return fmt.Errorf("no .feature spec found in specs/")
	}
	return nil
}

func validateTests() error {
	if !hasFileSuffix("tests/unit", ".test.ts") && !hasFileSuffix("tests/integration", ".test.ts") {
		return fmt.Errorf("no unit/integration tests found in tests/")
	}
	if !hasFileSuffix("tests/acceptance", ".steps.ts") {
		return fmt.Errorf("no acceptance step definitions found in tests/acceptance/")
	}
	return nil
}

func validateGatekeeper(feature string) error {
	report := fmt.Sprintf(".workflow/%s-gatekeeper.md", feature)
	if _, err := os.Stat(report); err != nil {
		return fmt.Errorf("gatekeeper report not found: %s", report)
	}
	return nil
}

func hasFileSuffix(dir, suffix string) bool {
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, suffix) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
