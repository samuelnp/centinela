package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// R3 tripwire. internal/config and internal/orchestration are both G2 LEAF
// packages and may not import each other, so neither can assert the invariant
// that binds them. cmd/ is the aggregator that already imports both and is
// where the two actually meet: the directive prints a tierModels id, and that
// same id flows into DefaultProfileForModel as an operator's driver_model.
//
// Changing tierModels without adding the new id to builtinModelCapability is
// SILENT — no error, no warning, just a different default enforcement profile.
// This test consumes the real table so that change cannot land quietly.
func TestBuiltinTierModelsAllClassify(t *testing.T) {
	ids := orchestration.TierModelIDs()
	if len(ids) == 0 {
		t.Fatal("TierModelIDs returned nothing — the tripwire would pass vacuously")
	}
	for _, id := range ids {
		class, ok := config.CapabilityClassFor(id, nil)
		if !ok || class == "" {
			t.Errorf("built-in model %q has no capability class: add it to "+
				"config.builtinModelCapability (never replace an existing key)", id)
			continue
		}
		if _, ok := config.DefaultProfileForModel(id, nil); !ok {
			t.Errorf("built-in model %q resolves no default enforcement profile", id)
		}
	}
}

// AC22 from the other direction: the map is a SUPERSET, so every id the table
// has ever shipped still classifies even though it is no longer a default.
func TestRetiredBuiltinIDsStillClassify(t *testing.T) {
	retired := []string{
		"claude-opus-4-7",
		"claude-sonnet-4-6",
		"claude-haiku-4-5-20251001",
	}
	for _, id := range retired {
		if _, ok := config.DefaultProfileForModel(id, nil); !ok {
			t.Errorf("retired pin %q lost its default profile — the capability "+
				"map must only ever grow", id)
		}
	}
}
