package acceptance_test

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

// scriptedRunner returns a fixed outcome for every command.
type scriptedRunner struct{ out verify.RunOutcome }

func (s scriptedRunner) Run(_, _ string, _ time.Duration) verify.RunOutcome { return s.out }

func accCfg() *config.Config {
	c := &config.Config{}
	c.Validate.Commands = []string{"go test ./..."}
	c.Verify.TimeoutSeconds = 60
	c.Verify.CoverageTolerance = 0.001
	return c
}

func runClaim(t *testing.T, root string, out verify.RunOutcome, ev *evidence.RoleEvidence) verify.VerificationResult {
	t.Helper()
	return verify.Verify("feat", "tests", accCfg(), verify.Deps{
		Root:   root,
		Runner: scriptedRunner{out: out},
		Load:   func(string, orchestration.Role) (*evidence.RoleEvidence, error) { return ev, nil },
	})
}

// Scenario: Fabricated tests-pass claim is blocked.
func TestAcceptance_FabricatedTestsBlocked(t *testing.T) {
	res := runClaim(t, t.TempDir(), verify.RunOutcome{ExitCode: 1}, &evidence.RoleEvidence{Role: "qa-senior"})
	if !res.HasFailures() {
		t.Fatal("non-zero test exit should hard-block")
	}
}

// Scenario: Empty-stub output file is blocked.
func TestAcceptance_StubOutputBlocked(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "tests", "unit")
	_ = os.MkdirAll(p, 0o755)
	_ = os.WriteFile(filepath.Join(p, "foo_test.go"),
		[]byte("package p\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n"), 0o644)
	ev := &evidence.RoleEvidence{Role: "qa-senior", Outputs: []string{"tests/unit/foo_test.go"}}
	res := runClaim(t, root, verify.RunOutcome{ExitCode: 0}, ev)
	if !res.HasFailures() {
		t.Fatal("empty-stub test file should hard-block")
	}
}

// Scenario: Inflated coverage claim is blocked.
func TestAcceptance_CoverageOverclaimBlocked(t *testing.T) {
	measured := verify.RunOutcome{Output: "coverage: 78.0% of statements\n"}
	v := 92.0
	res := runClaim(t, t.TempDir(), measured, &evidence.RoleEvidence{Role: "qa-senior", Coverage: &v})
	if !res.HasFailures() {
		t.Fatal("92%% claim vs 78%% measured should hard-block")
	}
}
