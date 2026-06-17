package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/enforcement-profiles.feature

// Scenario: An unconfigured project keeps today's behavior (default strict)
func TestEP_UnconfiguredKeepsStrictBehavior(t *testing.T) {
	t.Chdir(t.TempDir()) // no centinela.toml present
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	prof := workflow.EffectiveProfile(nil, cfg)
	if prof != config.ProfileStrict {
		t.Fatalf("effective profile = %q, want strict", prof)
	}
	knobs := config.ProfileDefaults(prof)
	if !knobs.StepGating {
		t.Fatal("step-gating must be on")
	}
	if knobs.ConfirmationMode != config.ConfirmEveryStep {
		t.Fatalf("confirmation mode = %q, want every_step", knobs.ConfirmationMode)
	}
	if !knobs.RequireSubagentEvidence {
		t.Fatal("subagent orchestration evidence must be required")
	}
}

// Scenario: An unknown profile value is rejected at config load
func TestEP_UnknownProfileRejectedAtLoad(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(config.Filename, []byte("[workflow]\nenforcement_profile=\"turbo\"\n"), 0644) //nolint:errcheck
	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "enforcement_profile") {
		t.Fatalf("expected load to fail naming enforcement_profile, got %v", err)
	}
}
