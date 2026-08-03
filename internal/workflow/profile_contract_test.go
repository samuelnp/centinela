package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func pinned() *Workflow { return &Workflow{ProfileContract: ProfileContractGuidedDefault} }

// TestUsesGuidedDefault covers the pin predicate in both directions, including
// the nil receiver EffectiveProfile relies on.
func TestUsesGuidedDefault(t *testing.T) {
	if !pinned().UsesGuidedDefault() {
		t.Fatal("the shipped contract must be recognized")
	}
	for name, wf := range map[string]*Workflow{
		"nil workflow":   nil,
		"legacy (empty)": {},
		"other contract": {ProfileContract: "something-else-v9"},
	} {
		if wf.UsesGuidedDefault() {
			t.Fatalf("%s must NOT take the guided default", name)
		}
	}
}

// TestNewWithOrderPinsProfileContract: every workflow born after the flip
// carries the pin, whatever profile it starts under.
func TestNewWithOrderPinsProfileContract(t *testing.T) {
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		wf := NewWithOrder("f", DefaultStepOrder, profile)
		if wf.ProfileContract != ProfileContractGuidedDefault {
			t.Fatalf("profile %q: contract = %q, want the guided-default pin", profile, wf.ProfileContract)
		}
	}
	if !New("f").UsesGuidedDefault() {
		t.Fatal("New must pin the contract too")
	}
}

// TestEffectiveProfile_TailIsStateDated is the blast-radius guard: the flip
// reaches pinned workflows and NOTHING else.
func TestEffectiveProfile_TailIsStateDated(t *testing.T) {
	if got := EffectiveProfile(pinned(), &config.Config{}); got != config.ProfileGuided {
		t.Fatalf("pinned + zero config = %q, want guided", got)
	}
	if got := EffectiveProfile(&Workflow{}, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("unpinned (legacy/in-flight) + zero config = %q, want strict", got)
	}
	if got := EffectiveProfile(nil, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("workflow-less caller = %q, want strict", got)
	}
	if got := EffectiveProfile(pinned(), nil); got != config.ProfileGuided {
		t.Fatalf("pinned + nil config = %q, want guided", got)
	}
}

// TestEffectiveProfile_PinLosesToEveryExplicitTier: the new tail is the LAST
// word, so all three higher tiers still outrank it.
func TestEffectiveProfile_PinLosesToEveryExplicitTier(t *testing.T) {
	tier1 := pinned()
	tier1.EnforcementProfile = config.ProfileStrict
	if got := EffectiveProfile(tier1, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("tier 1 --profile strict = %q, want strict", got)
	}
	if got := EffectiveProfile(pinned(), cfgWithProfile(config.ProfileStrict)); got != config.ProfileStrict {
		t.Fatalf("tier 2 explicit global strict = %q, want strict", got)
	}
	limited := pinned()
	limited.DriverModel = "haiku"
	if got := EffectiveProfile(limited, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("tier 3 limited driver = %q, want strict", got)
	}
}

// TestProfileProvenance_PinDoesNotMaskCapability: a capable driver model must
// still be reported as the SOURCE of guided, not the shipped default.
func TestProfileProvenance_PinDoesNotMaskCapability(t *testing.T) {
	capable := pinned()
	capable.DriverModel = "sonnet"
	profile, note := ProfileProvenance(capable, &config.Config{})
	if profile != config.ProfileGuided {
		t.Fatalf("profile = %q, want guided", profile)
	}
	if note != "driver: sonnet → capable" {
		t.Fatalf("note = %q, want the driver capability note", note)
	}
	if _, defaultNote := ProfileProvenance(pinned(), &config.Config{}); defaultNote == note {
		t.Fatal("capability-derived and default-derived guided must stay distinguishable")
	}
}
