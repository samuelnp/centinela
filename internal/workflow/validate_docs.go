package workflow

import (
	"fmt"
	"os"
)

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
	if _, err := os.Stat("docs/project-docs/index.html"); err != nil {
		return fmt.Errorf("documentation output not found: docs/project-docs/index.html")
	}
	return nil
}
