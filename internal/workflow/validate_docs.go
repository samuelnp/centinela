package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/orchestration"
)

const kbDir = "docs/project-docs/kb"

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
	if orchestration.IsUserFacingFeature(feature) {
		return validateDocsUserFacing(feature)
	}
	return validateDocsInternal(feature)
}

// validateDocsUserFacing keeps the full knowledge-base contract: the portal,
// the per-feature markdown guide, and its rendered page must all exist.
func validateDocsUserFacing(feature string) error {
	if _, err := os.Stat("docs/project-docs/index.html"); err != nil {
		return fmt.Errorf("documentation output not found: docs/project-docs/index.html")
	}
	kbMD := filepath.Join(kbDir, feature+".md")
	if _, err := os.Stat(kbMD); err != nil {
		return fmt.Errorf("knowledge base markdown missing for %q: %s (write a plain-language end-user guide with sections: What it does, When you'd use it, How it behaves)", feature, kbMD)
	}
	kbHTML := filepath.Join(kbDir, feature+".html")
	if _, err := os.Stat(kbHTML); err != nil {
		return fmt.Errorf("knowledge base page missing for %q: %s (run: centinela docs generate)", feature, kbHTML)
	}
	return nil
}

// validateDocsInternal requires only a one-line changelog entry; no
// knowledge-base guide or portal regeneration is needed for internal features.
func validateDocsInternal(feature string) error {
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
