// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: Evidence shape and scaffold size are unchanged by this feature
func TestTD_EvidenceShapeAndScaffoldSizeUnchanged(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteWorkflow(t, dir, "token-diet", "plan", "planner-v1", false)
	tdFeatureDocs(t, dir, "token-diet")

	// Per-role .md + .json evidence pairs are still both required under the
	// strict profile: init must still write both files.
	tdRun(t, bin, dir, "evidence", "init", "token-diet", "planner")
	for _, ext := range []string{".json", ".md"} {
		p := filepath.Join(dir, ".workflow", "token-diet-planner"+ext)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("evidence pair must still include %s: %v", ext, err)
		}
	}

	// WS3.5 (scaffold slimming) is out of scope: every archetype guide the
	// scaffold ships today must still be present.
	root := repoRoot(t)
	scaffoldDir := filepath.Join(root, "internal", "scaffold", "assets", "docs", "architecture")
	entries, err := os.ReadDir(scaffoldDir)
	if err != nil {
		t.Fatalf("scaffold architecture dir: %v", err)
	}
	guides := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			guides++
		}
	}
	const minGuides = 20 // this feature touches none of them; a real slimming pass would
	if guides < minGuides {
		t.Fatalf("scaffold architecture guides = %d, want at least %d (WS3.5 not attempted here)", guides, minGuides)
	}
}
