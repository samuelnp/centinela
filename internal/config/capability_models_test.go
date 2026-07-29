package config

import "testing"

// R3, the silent risk: an orchestration tierModels value that is NOT a key in
// builtinModelCapability returns ok=false from DefaultProfileForModel, which
// disengages the capability tier and changes an operator's default enforcement
// profile with no error at all.
//
// config is a LEAF layer (G2: allow = []), so this package must not import
// internal/orchestration — not even from a _test.go, since the import-graph
// gate reads TestImports too. The ids below are therefore literals, mirrored
// from orchestration.tierModels. The mechanical tripwire that feeds the REAL
// table through this map lives in cmd/centinela (the aggregator layer, which
// may import both leaves): see TestBuiltinTierModelsAllClassify.
func TestEveryBuiltinTierModelLiteralClassifies(t *testing.T) {
	builtins := []string{
		"opus", "anthropic/claude-opus-4-7",
		"sonnet", "anthropic/claude-sonnet-4-6",
		"haiku", "anthropic/claude-haiku-4-5",
	}
	for _, id := range builtins {
		class, ok := CapabilityClassFor(id, nil)
		if !ok || class == "" {
			t.Errorf("built-in tier model %q has no capability class", id)
		}
		if _, ok := DefaultProfileForModel(id, nil); !ok {
			t.Errorf("built-in tier model %q resolves no default profile", id)
		}
	}
}

// AC22: the map is a strict SUPERSET — every alias AND every retired pin keeps
// its class, so no operator's driver_model silently loses its default profile.
func TestCapabilityClassSupersetCoversAliasesAndLegacyPins(t *testing.T) {
	cases := map[string]string{
		// New family aliases.
		"opus":   CapabilityFrontier,
		"sonnet": CapabilityCapable,
		"haiku":  CapabilityLimited,
		// Provider-qualified opencode ids.
		"anthropic/claude-opus-4-7":   CapabilityFrontier,
		"anthropic/claude-sonnet-4-6": CapabilityCapable,
		"anthropic/claude-haiku-4-5":  CapabilityLimited,
		// Retired pins an operator's centinela.toml may still name.
		"claude-opus-4-7":           CapabilityFrontier,
		"claude-sonnet-4-6":         CapabilityCapable,
		"claude-haiku-4-5-20251001": CapabilityLimited,
		"claude-haiku-4-5":          CapabilityLimited,
	}
	for id, want := range cases {
		got, ok := CapabilityClassFor(id, nil)
		if !ok || got != want {
			t.Errorf("CapabilityClassFor(%q) = (%q, %v), want (%q, true)", id, got, ok, want)
		}
	}
}

// E25: a retired pin resolves to the same default profile as its alias twin.
func TestRetiredPinKeepsItsDefaultProfile(t *testing.T) {
	pairs := [][2]string{
		{"claude-opus-4-7", "opus"},
		{"claude-sonnet-4-6", "sonnet"},
		{"claude-haiku-4-5-20251001", "haiku"},
	}
	for _, pair := range pairs {
		legacy, lok := DefaultProfileForModel(pair[0], nil)
		alias, aok := DefaultProfileForModel(pair[1], nil)
		if !lok || !aok {
			t.Fatalf("%v: both ids must classify, got ok=(%v,%v)", pair, lok, aok)
		}
		if legacy != alias {
			t.Errorf("%q → %q but %q → %q; the pin lost its profile", pair[0], legacy, pair[1], alias)
		}
	}
	// The concrete expectations the capability tier was built on.
	for id, want := range map[string]string{"opus": ProfileOutcome, "sonnet": ProfileGuided, "haiku": ProfileStrict} {
		if got, _ := DefaultProfileForModel(id, nil); got != want {
			t.Errorf("DefaultProfileForModel(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestUnknownModelStillHasNoClass(t *testing.T) {
	if _, ok := CapabilityClassFor("some/unknown-model", nil); ok {
		t.Error("an unknown model id must NOT classify — the map is a superset, not a catch-all")
	}
}
