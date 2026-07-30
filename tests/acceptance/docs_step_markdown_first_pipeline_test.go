// Acceptance: specs/docs-step-markdown-first.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: The HTML pipeline is gone
// `centinela docs generate` (and `docs validate`) must fail loudly as unknown
// commands — never print help with exit 0.
func TestDSMFHTMLPipelineCommandsGone(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	for _, sub := range []string{"generate", "validate"} {
		out, code := runCent(t, bin, dir, "docs", sub)
		if code == 0 {
			t.Fatalf("docs %s must exit non-zero as an unknown command:\n%s", sub, out)
		}
		if !strings.Contains(out, "unknown command") {
			t.Fatalf("docs %s must report an unknown command:\n%s", sub, out)
		}
	}
}

// Scenario: The HTML pipeline is gone
// No code path regenerates docs/project-docs/index.html at merge time: the
// whole tree carries no reference to the portal-regen seam or docgen package.
func TestDSMFNoMergeTimePortalRegenCodePath(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "internal", "docgen")); !os.IsNotExist(err) {
		t.Fatalf("internal/docgen must not exist (stat err=%v)", err)
	}
	src, err := os.ReadFile(filepath.Join(root, "cmd", "centinela", "merge.go"))
	if err != nil {
		t.Fatalf("read merge.go: %v", err)
	}
	for _, banned := range []string{"docsPortalRegen", "docgen", "project-docs"} {
		if strings.Contains(string(src), banned) {
			t.Fatalf("merge.go still references %q", banned)
		}
	}
}
