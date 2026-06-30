package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestGatherEvidenceSurfacesReadError covers the readOptional/gatherEvidence
// error branch: a brief source that is a directory is a genuine I/O fault, not
// a graceful "missing" — both gatherEvidence and buildPRBody must surface it.
func TestGatherEvidenceSurfacesReadError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join("docs", "features", "feat.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := gatherEvidence("feat"); err == nil {
		t.Fatal("a directory brief should surface a read error")
	}
	if _, _, err := buildPRBody("feat"); err == nil {
		t.Fatal("buildPRBody must propagate the gatherEvidence error")
	}
}

// TestVerifyRootDetectsWorktree covers the worktree branch of verifyRoot.
func TestVerifyRootDetectsWorktree(t *testing.T) {
	wt := filepath.Join(t.TempDir(), ".worktrees", "feat")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(wt)
	if root := verifyRoot(); !strings.Contains(filepath.ToSlash(root), ".worktrees/feat") {
		t.Fatalf("verifyRoot should resolve the worktree root, got %q", root)
	}
}

// TestEmitCostWarningNoActiveWorkflow covers the wf==nil branch: cost is active
// but there is no active feature, so the soft gate stays silent.
func TestEmitCostWarningNoActiveWorkflow(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("centinela.toml",
		[]byte("[cost]\nenabled=true\nstep_token_budget=1000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	out := capture(t, func() error { emitCostWarning(cfg); return nil })
	if strings.TrimSpace(out) != "" {
		t.Fatalf("no active workflow should be silent, got %q", out)
	}
}

// TestRunRoadmapIterateWritesMarker covers the success path: it writes the
// suppression marker and reports it.
func TestRunRoadmapIterateWritesMarker(t *testing.T) {
	t.Chdir(t.TempDir())
	out := capture(t, func() error { return runRoadmapIterate(nil, nil) })
	if !strings.Contains(out, "marker written") {
		t.Fatalf("expected marker-written confirmation, got %q", out)
	}
}
