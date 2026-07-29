// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
)

const upsScaffoldMirrorPath = "internal/scaffold/assets/docs/architecture/planner-prompt.md"

// Scenario: The planner prompt is mirrored byte-identically in the scaffold tree
func TestUPS_PlannerPromptScaffoldMirrorByteIdentical(t *testing.T) {
	root := repoRoot(t)
	src, err := os.ReadFile(filepath.Join(root, upsPlannerPromptPath))
	if err != nil {
		t.Fatalf("source planner-prompt.md missing: %v", err)
	}
	mirror, err := os.ReadFile(filepath.Join(root, upsScaffoldMirrorPath))
	if err != nil {
		t.Fatalf("scaffold mirror missing: %v", err)
	}
	if string(src) != string(mirror) {
		t.Fatal("planner-prompt.md and its scaffold mirror must be byte-identical")
	}
	for _, gone := range []string{
		"internal/scaffold/assets/docs/architecture/big-thinker-prompt.md",
		"internal/scaffold/assets/docs/architecture/feature-specialist-prompt.md",
	} {
		if _, err := os.Stat(filepath.Join(root, gone)); !os.IsNotExist(err) {
			t.Fatalf("%s must no longer exist", gone)
		}
	}
}
