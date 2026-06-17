package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/githooks"
	"github.com/samuelnp/centinela/internal/ui"
)

// Acceptance spec: specs/precommit-and-pr-gate.feature
//
// Tests-tier unit coverage drives the public layer: the deterministic Markdown
// renderer over crafted gates.Result values, and the githooks Install/Uninstall
// round-trip in a temp dir.

func TestPrGate_MarkdownDeterministicAndMarked(t *testing.T) {
	results := []gates.Result{
		{Name: "G1: File Size", Status: gates.Fail, Message: "1 file over 100 lines",
			Details: []string{"internal/x.go (142 lines)"}},
		{Name: "import_graph", Status: gates.Pass, Message: "no forbidden edges"},
	}
	first := ui.RenderGatesMarkdown(results)
	second := ui.RenderGatesMarkdown(results)
	if first != second {
		t.Fatal("pr-gate markdown must be byte-identical across runs over identical input")
	}
	if !strings.HasPrefix(first, ui.MarkdownMarker) {
		t.Fatalf("markdown must lead with the CI marker: %q", first)
	}
	if !strings.Contains(first, "❌") || !strings.Contains(first, "internal/x.go (142 lines)") {
		t.Fatalf("failing gate + details must appear: %q", first)
	}
}

func TestGithooks_InstallUninstallRoundTrip(t *testing.T) {
	dir := t.TempDir()
	changed, err := githooks.Install(dir)
	if err != nil || !changed {
		t.Fatalf("install must succeed and change, got changed=%v err=%v", changed, err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "pre-commit"))
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if !strings.Contains(string(body), "centinela precommit") {
		t.Fatalf("hook must call centinela precommit: %q", body)
	}
	removed, err := githooks.Uninstall(dir)
	if err != nil || !removed {
		t.Fatalf("uninstall must remove the block, got changed=%v err=%v", removed, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pre-commit")); !os.IsNotExist(err) {
		t.Fatalf("centinela-only hook must be deleted on uninstall, stat err=%v", err)
	}
}
