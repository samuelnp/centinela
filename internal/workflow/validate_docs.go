package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
	return validateChangelog(feature)
}

// validateChangelog requires a non-empty changelog entry for every feature.
// The real-updated-doc-file rule for user-facing features lives in
// orchestration evidence validation (documentation-specialist outputs), not
// here.
func validateChangelog(feature string) error {
	path := filepath.Join(WorkflowDir, feature+"-changelog.md")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("changelog entry missing for %q: %s (write a one-line summary, e.g. via: centinela artifact new %s changelog)", feature, path, feature)
	}
	defer f.Close() //nolint:errcheck // read-only handle
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			return nil
		}
	}
	return fmt.Errorf("changelog entry is empty for %q: %s (write a one-line summary of the change)", feature, path)
}
