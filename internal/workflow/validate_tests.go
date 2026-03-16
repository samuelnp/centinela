package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

func validateTests(cfg *config.Config) error {
	suffixes := cfg.Workflow.TestSuffixes
	acceptance := cfg.Workflow.AcceptanceSuffix

	if !hasUnitOrIntegrationTests(suffixes) {
		return fmt.Errorf("no unit/integration tests found in tests/")
	}
	if !hasAcceptanceTests(acceptance) {
		return fmt.Errorf("no acceptance step definitions found in tests/acceptance/")
	}
	return nil
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

func hasAcceptanceTests(suffix string) bool {
	if suffix == "" {
		return hasAnyFile("tests/acceptance")
	}
	return hasFileSuffix("tests/acceptance", suffix)
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

func hasAnyFile(dir string) bool {
	found := false
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if !d.IsDir() {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
