package calibration

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// capCfg maps the three spec model ids to their classes so the unit tests do not
// depend on the exact built-in id spellings.
func capCfg() *config.Config {
	c := &config.Config{}
	c.Orchestration.Capabilities = map[string]string{
		"claude-opus-4-7":   config.CapabilityFrontier,
		"claude-sonnet-4-6": config.CapabilityCapable,
		"claude-haiku-4-5":  config.CapabilityLimited,
	}
	return c
}

// rated builds a FrictionStats with the given advances/rework, computing rate.
func rated(adv, rew int) FrictionStats {
	s := FrictionStats{Advances: adv, Rework: rew}
	if adv > 0 {
		s.Rate = float64(rew) / float64(adv)
		s.HasRate = true
	}
	return s
}

// TestClassifyUnclassified — no capability class → Unclassified/None, no profile.
func TestClassifyUnclassified(t *testing.T) {
	v, r, rp, class, cur := classify("local/unknown", rated(5, 5), nil)
	if v != Unclassified || r != None || rp != "" || class != "" || cur != "" {
		t.Fatalf("unclassified wrong: %v %v %q %q %q", v, r, rp, class, cur)
	}
}

// TestClassifyInsufficientAdvances — Advances<minAdvances → Keep regardless of rate.
func TestClassifyInsufficientAdvances(t *testing.T) {
	v, r, rp, _, _ := classify("claude-haiku-4-5", rated(2, 5), capCfg())
	if v != WellCalibrated || r != Keep || rp != config.ProfileStrict {
		t.Fatalf("insufficient-advances wrong: %v %v %q", v, r, rp)
	}
	v2, r2, _, _, _ := classify("claude-haiku-4-5", FrictionStats{GateFailures: 4, Rework: 4}, capCfg())
	if v2 != WellCalibrated || r2 != Keep {
		t.Fatalf("zero-advance classify wrong: %v %v", v2, r2)
	}
}

// TestClassifyTighten — high friction (Rate≥1.0) under a tightenable profile.
func TestClassifyTighten(t *testing.T) {
	v, r, rp, _, cur := classify("claude-sonnet-4-6", rated(3, 3), capCfg()) // rate 1.0
	if v != Undergoverned || r != Tighten || rp != config.ProfileStrict || cur != config.ProfileGuided {
		t.Fatalf("tighten wrong: %v %v %q cur=%q", v, r, rp, cur)
	}
}

// TestClassifyTightenMaxed — high friction already at strict → Keep.
func TestClassifyTightenMaxed(t *testing.T) {
	v, r, rp, _, _ := classify("claude-haiku-4-5", rated(3, 3), capCfg())
	if v != WellCalibrated || r != Keep || rp != config.ProfileStrict {
		t.Fatalf("maxed-strict wrong: %v %v %q", v, r, rp)
	}
}

// TestClassifyLoosen — low friction (Rate≤0.25) under a loosenable profile, incl.
// the Rate=0.0 case.
func TestClassifyLoosen(t *testing.T) {
	v, r, rp, _, _ := classify("claude-haiku-4-5", rated(4, 1), capCfg()) // rate 0.25
	if v != Overgoverned || r != Loosen || rp != config.ProfileGuided {
		t.Fatalf("loosen wrong: %v %v %q", v, r, rp)
	}
	v2, r2, _, _, _ := classify("claude-haiku-4-5", rated(5, 0), capCfg()) // rate 0.0
	if v2 != Overgoverned || r2 != Loosen {
		t.Fatalf("rate-0 loosen wrong: %v %v", v2, r2)
	}
}

// TestClassifyLoosenMaxed / Between — low friction at outcome → Keep; mid → Keep.
func TestClassifyLoosenMaxedAndBetween(t *testing.T) {
	v, r, rp, _, _ := classify("claude-opus-4-7", rated(5, 1), capCfg()) // rate 0.2 at outcome
	if v != WellCalibrated || r != Keep || rp != config.ProfileOutcome {
		t.Fatalf("maxed-outcome wrong: %v %v %q", v, r, rp)
	}
	v2, r2, _, _, _ := classify("claude-sonnet-4-6", rated(4, 2), capCfg()) // rate 0.5
	if v2 != WellCalibrated || r2 != Keep {
		t.Fatalf("between wrong: %v %v", v2, r2)
	}
}
