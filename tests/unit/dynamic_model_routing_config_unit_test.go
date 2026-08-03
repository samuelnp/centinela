package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// Parity: every role slug the DOMAIN advertises must be accepted as an
// [orchestration.floors] key, and every tier as a value. The config leaf keeps
// its own local allow-lists (it may not import orchestration), so only a
// cross-package test can keep the two vocabularies from drifting.
func TestFloorsAllowListParity_RolesAndTiersAccepted(t *testing.T) {
	for _, role := range orchestration.AllowedRoleSlugs() {
		if _, err := loadTempConfig(t, "[orchestration.floors]\n"+role+" = \"balanced\"\n"); err != nil {
			t.Errorf("role slug %q from AllowedRoleSlugs() rejected as a floors key: %v", role, err)
		}
	}
	for _, tier := range orchestration.AllowedTiers() {
		toml := "[orchestration.floors]\nqa-senior = \"" + string(tier) + "\"\n"
		if _, err := loadTempConfig(t, toml); err != nil {
			t.Errorf("tier %q from AllowedTiers() rejected as a floors value: %v", tier, err)
		}
	}
}

func TestFloorsValidation_UnknownRoleAndTierRejected(t *testing.T) {
	_, err := loadTempConfig(t, "[orchestration.floors]\nwizard = \"fast\"\n")
	if err == nil || !strings.Contains(err.Error(), "wizard") {
		t.Fatalf("an unknown floors role key must be refused naming it, got %v", err)
	}
	_, err = loadTempConfig(t, "[orchestration.floors]\nqa-senior = \"turbo\"\n")
	if err == nil || !strings.Contains(err.Error(), "reasoning, balanced, fast") {
		t.Fatalf("an unknown floors tier must be refused listing the tiers, got %v", err)
	}
}

// E9 — one vocabulary, two validators: [orchestration.floors] and
// [orchestration.models] must give the SAME answer for the same key, in every
// casing. Accepting "Gatekeeper" as a floor while rejecting it as a model taught
// operators a spelling the rest of the file refuses — and silently APPLIED it.
func TestFloorsAndModels_KeyCasingParity(t *testing.T) {
	for _, role := range orchestration.AllowedRoleSlugs() {
		for _, key := range []string{role, strings.ToUpper(role), titleFirst(role)} {
			_, floorsErr := loadTempConfig(t, "[orchestration.floors]\n"+key+" = \"balanced\"\n")
			_, modelsErr := loadTempConfig(t, "[orchestration.models]\n"+key+" = \"balanced\"\n")
			if (floorsErr == nil) != (modelsErr == nil) {
				t.Errorf("key %q disagrees: floors=%v models=%v", key, floorsErr, modelsErr)
			}
		}
	}
}

// The routing_mode key itself: absent, both values, and a typo.
func TestRoutingMode_AcceptedValuesAndTypo(t *testing.T) {
	for _, mode := range []string{"static", "dynamic", "Dynamic", " dynamic "} {
		cfg, err := loadTempConfig(t, "[orchestration]\nrouting_mode = \""+mode+"\"\n")
		if err != nil {
			t.Fatalf("routing_mode %q must load: %v", mode, err)
		}
		want := strings.EqualFold(strings.TrimSpace(mode), "dynamic")
		if config.DynamicRoutingEnabled(cfg) != want {
			t.Fatalf("routing_mode %q: dynamic=%v, want %v", mode, !want, want)
		}
	}
	_, err := loadTempConfig(t, "[orchestration]\nrouting_mode = \"dinamyc\"\n")
	if err == nil || !strings.Contains(err.Error(), "routing_mode") {
		t.Fatalf("a typo must fail at load naming the key, got %v", err)
	}
}

func titleFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
