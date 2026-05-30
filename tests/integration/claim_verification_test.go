package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/verify"
)

// fakeRunner lets the gate-wiring test inject a failing test run without
// shelling out.
type fakeRunner struct{ exit int }

func (f fakeRunner) Run(_ string, _ string, _ time.Duration) verify.RunOutcome {
	return verify.RunOutcome{ExitCode: f.exit}
}

func gateCfg() *config.Config {
	c := &config.Config{}
	c.Validate.Commands = []string{"go test ./..."}
	c.Verify.TimeoutSeconds = 60
	c.Verify.CoverageTolerance = 0.001
	return c
}

// TestCompleteGateBlocksFabricatedClaim wires the verify domain the way the
// complete gate does: a "tests pass" claim with a non-zero test run hard-blocks.
func TestCompleteGateBlocksFabricatedClaim(t *testing.T) {
	ev := &evidence.RoleEvidence{Feature: "fab", Role: "qa-senior", Status: "done"}
	deps := verify.Deps{
		Root:   t.TempDir(),
		Runner: fakeRunner{exit: 1},
		Load: func(string, orchestration.Role) (*evidence.RoleEvidence, error) {
			return ev, nil
		},
	}
	res := verify.Verify("fab", "tests", gateCfg(), deps)
	if !res.HasFailures() {
		t.Fatal("fabricated tests-pass claim must produce a blocking failure")
	}
}

// TestVerifyResolvesWorktreeRoot confirms verify reads evidence and tests from
// the resolved root, not the process CWD.
func TestVerifyResolvesWorktreeRoot(t *testing.T) {
	root := t.TempDir()
	wtTests := filepath.Join(root, "tests", "unit")
	if err := os.MkdirAll(wtTests, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(wtTests, "x_test.go"),
		[]byte("package p\nfunc TestMissingEvidenceFiles(t *testing.T){}\n"), 0o644)
	ev := &evidence.RoleEvidence{Feature: "wt", Role: "qa-senior", EdgeCases: []string{"missing evidence files"}}
	deps := verify.Deps{
		Root:   root,
		Runner: fakeRunner{exit: 0},
		Load:   func(string, orchestration.Role) (*evidence.RoleEvidence, error) { return ev, nil },
	}
	res := verify.Verify("wt", "tests", gateCfg(), deps)
	if res.HasWarnings() {
		t.Fatalf("edge case present in worktree tests should not warn: %+v", res.Checks)
	}
}
