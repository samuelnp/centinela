package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgWithProfile(p string) *config.Config {
	c := &config.Config{}
	c.Workflow.EnforcementProfile = p
	return c
}

// Precedence: per-feature override > global config > strict default.
func TestEffectiveProfile_Precedence(t *testing.T) {
	wf := &Workflow{EnforcementProfile: config.ProfileOutcome}
	if got := EffectiveProfile(wf, cfgWithProfile(config.ProfileGuided)); got != config.ProfileOutcome {
		t.Fatalf("per-feature override must win, got %q", got)
	}
	bare := &Workflow{}
	if got := EffectiveProfile(bare, cfgWithProfile(config.ProfileGuided)); got != config.ProfileGuided {
		t.Fatalf("global config must win when no per-feature value, got %q", got)
	}
	if got := EffectiveProfile(bare, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("unconfigured must default to strict, got %q", got)
	}
}

func TestEffectiveProfile_NilInputs(t *testing.T) {
	if got := EffectiveProfile(nil, nil); got != config.ProfileStrict {
		t.Fatalf("nil wf+cfg must default to strict, got %q", got)
	}
	if got := EffectiveProfile(nil, cfgWithProfile(config.ProfileOutcome)); got != config.ProfileOutcome {
		t.Fatalf("nil wf must fall to global, got %q", got)
	}
}

func TestDisplayProfile(t *testing.T) {
	if DisplayProfile(nil) != config.ProfileStrict {
		t.Fatal("nil workflow must display strict")
	}
	if DisplayProfile(&Workflow{}) != config.ProfileStrict {
		t.Fatal("empty pinned profile must display strict")
	}
	if DisplayProfile(&Workflow{EnforcementProfile: config.ProfileGuided}) != config.ProfileGuided {
		t.Fatal("pinned guided must display guided")
	}
}
