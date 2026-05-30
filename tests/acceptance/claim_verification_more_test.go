package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/verify"
)

// Scenario: Edge case with no corresponding test emits a warning but does not
// hard-block.
func TestAcceptance_UnmappedEdgeWarnsNotBlocks(t *testing.T) {
	ev := &evidence.RoleEvidence{Role: "qa-senior", EdgeCases: []string{"galaxy supernova explodes"}}
	res := runClaim(t, t.TempDir(), verify.RunOutcome{ExitCode: 0}, ev)
	if !res.HasWarnings() {
		t.Fatal("unmatched edge case should WARN")
	}
	if res.HasFailures() {
		t.Fatal("edge-case warning alone must NOT hard-block completion")
	}
}

// Scenario: Honest evidence verifies green and completes unchanged.
func TestAcceptance_HonestEvidenceGreen(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "tests", "unit")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "real_test.go"),
		[]byte("package p\nimport \"testing\"\nfunc TestMappedEdge(t *testing.T){ if false { t.Fatal(\"x\") } }\n"), 0o644)
	v := 84.95
	ev := &evidence.RoleEvidence{
		Role:      "qa-senior",
		Coverage:  &v,
		Outputs:   []string{"tests/unit/real_test.go"},
		EdgeCases: []string{"mapped edge"},
	}
	res := runClaim(t, root, verify.RunOutcome{Output: "coverage: 85.0% of statements\n"}, ev)
	if res.HasFailures() || res.HasWarnings() {
		t.Fatalf("honest evidence should verify clean: %+v", res.Checks)
	}
}

// Scenario: No evidence files for a step reports skip and does not block.
func TestAcceptance_NoEvidenceSkips(t *testing.T) {
	res := verify.Verify("fresh", "tests", accCfg(), verify.Deps{
		Root:   t.TempDir(),
		Runner: scriptedRunner{out: verify.RunOutcome{ExitCode: 0}},
		Load:   func(string, orchestration.Role) (*evidence.RoleEvidence, error) { return nil, os.ErrNotExist },
	})
	if res.HasFailures() {
		t.Fatal("absent evidence must not block")
	}
}
