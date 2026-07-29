// Integration coverage for docs/plans/adversarial-validate-verifier.md.
//
// The colocated unit tests in internal/workflow (validate_gatekeeper_test.go,
// validate_freshness_test.go, validate_freshness_stale_test.go) exercise this
// logic against a STUBBED treestate.Runner. This file drives the same
// assembled behavior — workflow.ValidateArtifacts + workflow.VerificationFresh
// — through a REAL git repo and the REAL exec-based Runner, so the wiring
// between the two packages is proven, not just each package in isolation.
package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/treestate"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestRealGitFreshnessRoundTrip exercises the single highest-value assertion
// in the whole feature (plan §6) through the real git Runner: a fresh stamp
// passes both gates, an uncommitted tracked edit stales it, and further
// .workflow/-only churn never does (D3a).
func TestRealGitFreshnessRoundTrip(t *testing.T) {
	dir := t.TempDir()
	avviGit(t, dir, "init", "-q", "-b", "main")
	avviGit(t, dir, "config", "user.email", "avvi@centinela.dev")
	avviGit(t, dir, "config", "user.name", "AVVI")
	mustWrite(t, dir, "src.go", "package x\n")
	avviGit(t, dir, "add", ".")
	avviGit(t, dir, "commit", "-q", "-m", "baseline")

	feature := "real-git-fresh"
	avviWorkflow(t, dir, feature)
	report := filepath.Join(dir, ".workflow", feature+"-gatekeeper.md")
	avviWriteVerification(t, report, "", "")

	// Exercise the REAL production pipeline: StampVerification -> gatereport
	// splice -> treestate.Stamp -> the real exec.Command git Runner.
	runner := treestate.NewExecRunner()
	if _, err := workflow.StampVerification(feature, dir, runner); err != nil {
		t.Fatalf("StampVerification: %v", err)
	}

	cfg := &config.Config{}
	avviChdir(t, dir, func() {
		if err := workflow.ValidateArtifacts(feature, "validate", cfg); err != nil {
			t.Fatalf("ValidateArtifacts must pass a grounded fresh report: %v", err)
		}
		if err := workflow.VerificationFresh(feature, ".", runner); err != nil {
			t.Fatalf("VerificationFresh must pass immediately after stamping: %v", err)
		}
	})

	mustWrite(t, dir, "src.go", "package x\n\nfunc changed() {}\n")
	avviChdir(t, dir, func() {
		if err := workflow.VerificationFresh(feature, ".", runner); err == nil {
			t.Fatal("an uncommitted tracked-file edit must stale a real-git freshness check")
		}
	})
	avviGit(t, dir, "checkout", "--", "src.go")

	mustWrite(t, dir, ".workflow/"+feature+"-scratch.json", `{"noise":true}`)
	avviChdir(t, dir, func() {
		if err := workflow.VerificationFresh(feature, ".", runner); err != nil {
			t.Fatalf(".workflow/-only churn must not stale a real-git freshness check: %v", err)
		}
	})
}
