// Acceptance: specs/right-size-docs-step.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rdsMergeSource reads cmd/centinela/merge.go from the repo root.
func rdsMergeSource(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "cmd", "centinela", "merge.go"))
	if err != nil {
		t.Fatalf("read merge.go: %v", err)
	}
	return string(data)
}

// Scenario: A clean merge regenerates the documentation portal
func TestRDSCleanMergeRegeneratesPortal(t *testing.T) {
	src := rdsMergeSource(t)
	// The clean-merge path must invoke the portal-regen seam.
	if !strings.Contains(src, "docsPortalRegen()") {
		t.Fatal("a clean merge must call docsPortalRegen() to refresh the portal")
	}
	if !strings.Contains(src, "docgen.Generate(") {
		t.Fatal("the regen seam must be wired to docgen.Generate")
	}
}

// Scenario: A portal regeneration failure does not fail a clean merge
func TestRDSPortalRegenFailureDoesNotFailMerge(t *testing.T) {
	src := rdsMergeSource(t)
	// On a regen error the merge prints a notice and continues (no return err).
	idx := strings.Index(src, "docsPortalRegen()")
	if idx < 0 {
		t.Fatal("docsPortalRegen call not found")
	}
	tail := src[idx:]
	if !strings.Contains(tail, "notice: portal regen skipped") {
		t.Fatal("a regen failure must emit a notice rather than fail the merge")
	}
	if strings.Contains(tail[:strings.Index(tail, "RenderSuccess")], "return err") {
		t.Fatal("a regen failure must not return an error from the merge")
	}
}
