package acceptance_test

// Acceptance: specs/deferred-findings-roadmap-capture.feature
// Scenario: All eight role prompts and their scaffold mirrors contain the Deferred Findings section byte-identically

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var dfrcPromptFiles = []string{
	"big-thinker-prompt.md",
	"feature-specialist-prompt.md",
	"senior-engineer-prompt.md",
	"qa-senior-prompt.md",
	"edge-case-tester-prompt.md",
	"ux-ui-specialist-prompt.md",
	"validation-specialist-prompt.md",
	"gatekeeper-prompt.md",
}

func dfrcRepoRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

// TestDfrc_PromptParityByteIdentical verifies each source prompt is byte-identical
// to its scaffold mirror, and both contain the Deferred Findings section.
func TestDfrc_PromptParityByteIdentical(t *testing.T) {
	root := dfrcRepoRoot(t)
	srcDir := filepath.Join(root, "docs", "architecture")
	mirrorDir := filepath.Join(root, "internal", "scaffold", "assets", "docs", "architecture")

	for _, name := range dfrcPromptFiles {
		name := name
		t.Run(name, func(t *testing.T) {
			srcPath := filepath.Join(srcDir, name)
			mirrorPath := filepath.Join(mirrorDir, name)

			srcData, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("read source prompt %s: %v", name, err)
			}
			mirrorData, err := os.ReadFile(mirrorPath)
			if err != nil {
				t.Fatalf("read mirror prompt %s: %v", name, err)
			}

			if !bytes.Equal(srcData, mirrorData) {
				t.Errorf("source and mirror are not byte-identical: %s", name)
			}
			if !strings.Contains(string(srcData), "Deferred Findings") {
				t.Errorf("source prompt %s missing 'Deferred Findings' section", name)
			}
			if !strings.Contains(string(srcData), "roadmap defer") {
				t.Errorf("source prompt %s missing 'roadmap defer' reference", name)
			}
		})
	}
}
