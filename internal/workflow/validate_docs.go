package workflow

import (
	"fmt"
	"os"
	"path/filepath"
)

const kbDir = "docs/project-docs/kb"

func validateDocsOutput(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature is required for docs validation")
	}
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
