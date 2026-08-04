package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestEffectiveProfile_TailNeedsARealLoad is the value-level fail-closed rule:
// the SAME pinned workflow answers guided under a config a successful Load
// produced and strict under one nobody loaded. This is what makes a swallowed
// config error safe at every surface at once, instead of one call site per round.
func TestEffectiveProfile_TailNeedsARealLoad(t *testing.T) {
	if got := EffectiveProfile(pinned(), loadedCfg(t)); got != config.ProfileGuided {
		t.Fatalf("a real load must reach the tail, got %q", got)
	}
	for name, cfg := range map[string]*config.Config{
		"nil config":        nil,
		"fabricated config": {},
	} {
		if got := EffectiveProfile(pinned(), cfg); got != config.ProfileStrict {
			t.Fatalf("%s must never reach the guided tail, got %q", name, got)
		}
	}
}
