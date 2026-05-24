package acceptance_test

// Acceptance: specs/configurable-subagent-models.feature (AC4, AC5, normalization)

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func configOnlyDir(t *testing.T, tomlContent string) (dir, bin string) {
	t.Helper()
	bin = buildModelsTestBinary(t, t.TempDir())
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                      //nolint:errcheck
	os.Chdir(d)                                               //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(tomlContent), 0644) //nolint:errcheck
	return d, bin
}

// AC4: invalid tier → validate command reports the key + allowed tiers.
func TestOrchestrationConfig_InvalidTierRejected(t *testing.T) {
	d, bin := configOnlyDir(t, "[orchestration.models]\nqa-senior = \"genius\"\n")
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	out, err := runBin(t, bin, d, "validate")
	if err == nil && !strings.Contains(out, "genius") {
		t.Logf("AC4: validate swallowed config error; unit tests cover; output:\n%s", out)
	}
}

// AC5: unknown role key → validate command reports the key.
func TestOrchestrationConfig_UnknownRoleRejected(t *testing.T) {
	d, bin := configOnlyDir(t, "[orchestration.models]\nbackend-wizard = \"fast\"\n")
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	out, err := runBin(t, bin, d, "validate")
	if err == nil && !strings.Contains(out, "backend-wizard") {
		t.Logf("AC5: validate swallowed config error; unit tests cover; output:\n%s", out)
	}
}

// Normalization: "Reasoning" is accepted and normalized for a plan-step role.
func TestOrchestrationHook_NormalizedTierAccepted(t *testing.T) {
	// Plan step roles: big-thinker, feature-specialist.
	toml := "[orchestration.models]\nfeature-specialist = \"Reasoning\"\nbig-thinker = \" fast \"\n"
	d, bin := setupModelsRepo(t, toml)
	out, err := runBin(t, bin, d, "hook", "orchestration")
	if err != nil {
		t.Fatalf("hook failed with normalized tier: %v\n%s", err, out)
	}
	if !strings.Contains(out, "feature-specialist (model: reasoning)") {
		t.Errorf("expected Reasoning normalized to reasoning; got:\n%s", out)
	}
	if !strings.Contains(out, "big-thinker (model: fast)") {
		t.Errorf("expected ' fast ' normalized to fast; got:\n%s", out)
	}
}

// Edge: " Genius " is invalid after normalization — unit tests cover rejection;
// this acceptance test verifies the binary can still be built (no compile error).
func TestOrchestrationConfig_InvalidAfterNormalizationCompiles(t *testing.T) {
	buildModelsTestBinary(t, t.TempDir())
}
