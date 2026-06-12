package config

import (
	"os"
	"strings"
	"testing"
)

// An explicitly-set unknown enforcement_profile must be rejected at Load (the
// error names the field), proving validation runs against the RAW value before
// applyDefaults normalizes it to strict.
func TestLoad_UnknownProfileRejected(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\nenforcement_profile=\"turbo\"\n"), 0644) //nolint:errcheck
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "enforcement_profile") {
		t.Fatalf("expected enforcement_profile rejection, got %v", err)
	}
}

func TestLoad_KnownProfileAccepted(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\nenforcement_profile=\"outcome\"\n"), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Workflow.EnforcementProfile != ProfileOutcome {
		t.Fatalf("profile = %q, want outcome", cfg.Workflow.EnforcementProfile)
	}
}

// applyDefaults overwrites an empty StepConfirmationMode with every_step. The
// RawStepConfirmationMode shadow must preserve the explicit-vs-defaulted signal:
// empty raw when unset, the literal value when set.
func TestLoad_RawStepConfirmationMode_DistinguishesExplicit(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\n"), 0644) //nolint:errcheck
	cfg, _ := Load()
	if cfg.Workflow.RawStepConfirmationMode != "" {
		t.Fatalf("unset mode must leave raw empty, got %q", cfg.Workflow.RawStepConfirmationMode)
	}
	if cfg.Workflow.StepConfirmationMode != ConfirmEveryStep {
		t.Fatalf("normalized mode should default to every_step, got %q", cfg.Workflow.StepConfirmationMode)
	}

	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\nstep_confirmation_mode=\"every_step\"\n"), 0644) //nolint:errcheck
	cfg2, _ := Load()
	if cfg2.Workflow.RawStepConfirmationMode != "every_step" {
		t.Fatalf("explicit mode must be captured raw, got %q", cfg2.Workflow.RawStepConfirmationMode)
	}
}
