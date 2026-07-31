package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
	return validateChangelog(feature)
}

// validateChangelog requires a real changelog entry for every feature: a
// non-empty first line that is no longer the scaffold `centinela artifact new
// <feature> changelog` writes. A stub is non-empty, so "non-blank" alone let
// the docs step complete on an entry nobody wrote.
//
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
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Only the ENTRY line is checked: later lines are free prose, and a
		// changelog that legitimately discusses the marker must stay writable.
		if orchestration.HasFillMarker(line) {
			return fmt.Errorf("changelog entry is still a template placeholder for %q: %s (replace the <FILL: ...> slots with a real one-line summary of the change)", feature, path)
		}
		return nil
	}
	return fmt.Errorf("changelog entry is empty for %q: %s (write a one-line summary of the change)", feature, path)
}
