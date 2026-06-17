package gates

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestClassifySecrets_FindingYieldsFail exercises AC1: retained finding -> Fail.
func TestClassifySecrets_FindingYieldsFail(t *testing.T) {
	p := writeReport(t, `[{"RuleID":"aws-key","File":"src/config.go"}]`)
	r := classifySecrets(p, "", nil, config.SecretsConfig{}, nil)
	if r.Status != Fail {
		t.Fatalf("finding must yield Fail, got %v", r.Status)
	}
	if len(r.Details) == 0 {
		t.Fatal("Fail result must carry details")
	}
}

// TestClassifySecrets_NoFindingsYieldsPass exercises AC2: clean -> Pass.
func TestClassifySecrets_NoFindingsYieldsPass(t *testing.T) {
	p := writeReport(t, `[]`)
	r := classifySecrets(p, "", nil, config.SecretsConfig{}, nil)
	if r.Status != Pass {
		t.Fatalf("no findings must yield Pass, got %v: %q", r.Status, r.Message)
	}
}

// TestClassifySecrets_MalformedJSONYieldsWarn exercises malformed-output edge:
// never a false Pass when JSON is unparseable.
func TestClassifySecrets_MalformedJSONYieldsWarn(t *testing.T) {
	p := writeReport(t, "not json")
	r := classifySecrets(p, "", nil, config.SecretsConfig{}, nil)
	if r.Status != Warn {
		t.Fatalf("malformed JSON must yield Warn, not %v", r.Status)
	}
	if r.Status == Pass {
		t.Fatal("malformed JSON must NEVER yield false Pass")
	}
}

// TestClassifySecrets_AllowlistExcludesOnlyFinding exercises AC6: allowlisted
// finding -> Pass (not Fail).
func TestClassifySecrets_AllowlistExcludesOnlyFinding(t *testing.T) {
	p := writeReport(t, `[{"RuleID":"generic-api-key","File":"config.go"}]`)
	cfg := config.SecretsConfig{Allowlist: []string{"generic-api-key"}}
	r := classifySecrets(p, "", nil, cfg, nil)
	if r.Status != Pass {
		t.Fatalf("allowlisted-only finding must yield Pass, got %v", r.Status)
	}
}

// TestClassifySecrets_LaunchFailureIsNotPass exercises the launch-failure path.
func TestClassifySecrets_LaunchFailureIsNotPass(t *testing.T) {
	r := classifySecrets("/no/such/file.json", "err", errScanTimeout, config.SecretsConfig{}, nil)
	if r.Status == Pass {
		t.Fatal("launch failure must not yield Pass")
	}
}
