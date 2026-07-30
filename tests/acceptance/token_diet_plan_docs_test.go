// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: The architecture docs and their scaffold mirrors state the new rule
func TestTD_ArchitectureDocsStateTheNewRuleAndMirrorByteIdentical(t *testing.T) {
	root := repoRoot(t)
	docs := []string{
		"docs/architecture/evidence-contract.md",
		"docs/architecture/planner-prompt.md",
		"docs/architecture/workflow-enforcement.md",
	}
	for _, rel := range docs {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		text := string(data)
		if strings.Contains(text, "docs/features/*.md") {
			t.Fatalf("%s must not require snapshotting every docs/features/*.md", rel)
		}
		if !strings.Contains(text, "Additional inputs are allowed") &&
			!strings.Contains(text, "more is fine") {
			t.Fatalf("%s must state that additional inputs are allowed", rel)
		}

		mirrorRel := "internal/scaffold/assets/" + rel
		mirror, err := os.ReadFile(filepath.Join(root, mirrorRel))
		if err != nil {
			t.Fatalf("mirror for %s: %v", rel, err)
		}
		if string(mirror) != text {
			t.Fatalf("%s and its scaffold mirror %s must be byte-identical", rel, mirrorRel)
		}
	}
}
