package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

func validateTests(feature string, cfg *config.Config) error {
	suffixes := cfg.Workflow.TestSuffixes
	acceptance := cfg.Workflow.AcceptanceSuffix

	if !hasUnitOrIntegrationTests(suffixes) {
		return fmt.Errorf("no unit/integration tests found in tests/")
	}
	if !hasAcceptanceTests(acceptance) {
		return fmt.Errorf("no executable acceptance test artifacts found in tests/acceptance/")
	}
	if !hasAcceptanceExecutionCommand(cfg.Validate.Commands) {
		return fmt.Errorf("validate.commands must include a command that executes acceptance tests")
	}
	if !hasEdgeCaseReport(feature) {
		return fmt.Errorf("edge-case report missing: .workflow/%s-edge-cases.md", feature)
	}
	return nil
}

func hasEdgeCaseReport(feature string) bool {
	if feature == "" {
		return false
	}
	path := fmt.Sprintf(".workflow/%s-edge-cases.md", feature)
	_, err := os.Stat(path)
	return err == nil
}

func hasUnitOrIntegrationTests(suffixes []string) bool {
	if len(suffixes) == 0 {
		return hasAnyFile("tests/unit") || hasAnyFile("tests/integration")
	}
	for _, s := range suffixes {
		if hasFileSuffix("tests/unit", s) || hasFileSuffix("tests/integration", s) {
			return true
		}
	}
	return false
}

func hasAcceptanceTests(suffix string) bool { return hasExecutableAcceptanceTests(suffix) }

func hasFileSuffix(dir, suffix string) bool {
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, suffix) && isRealTestArtifact(path) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func hasAnyFile(dir string) bool {
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if !d.IsDir() && isRealTestArtifact(path) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func isRealTestArtifact(path string) bool {
	name := filepath.Base(path)
	return name != ".gitkeep" && !strings.HasPrefix(name, ".")
}
