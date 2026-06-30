package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/ui"
)

// TestBrownfieldDetectionWiresToEnrichDirective exercises the real detector and
// the real renderer together: a repo with a manifest is classified brownfield,
// and the brownfield panel carries the analyze/synthesize/enrich guidance the
// hook injects. This is the cross-package contract the CLI relies on.
func TestBrownfieldDetectionWiresToEnrichDirective(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !analyze.HasSource(dir) {
		t.Fatal("repo with go.mod must be detected as brownfield")
	}

	panel := ui.RenderBrownfieldSetupNeeded()
	for _, want := range []string{"analyze", "synthesize", "ENRICH", "**Project Stage:** existing"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("brownfield panel missing %q", want)
		}
	}
}

// TestEmptyRepoStaysGreenfield: with no manifest and only an empty src/, the
// detector returns false so the hook keeps the greenfield setup path.
func TestEmptyRepoStaysGreenfield(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if analyze.HasSource(dir) {
		t.Fatal("empty repo (only an empty src/) must read as greenfield")
	}
}
