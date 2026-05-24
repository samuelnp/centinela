package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestDefaultTierForRole_SevenStepRoles(t *testing.T) {
	cases := []struct {
		role orchestration.Role
		want orchestration.Tier
	}{
		{orchestration.RoleBigThinker, orchestration.TierReasoning},
		{orchestration.RoleSeniorEngineer, orchestration.TierReasoning},
		{orchestration.RoleFeatureSpecial, orchestration.TierBalanced},
		{orchestration.RoleQASeniorEngineer, orchestration.TierBalanced},
		{orchestration.RoleUXUISpecialist, orchestration.TierBalanced},
		{orchestration.RoleDocsSpecialist, orchestration.TierFast},
		{orchestration.RoleValidationSpec, orchestration.TierFast},
	}
	for _, tc := range cases {
		if got := orchestration.DefaultTierForRole(tc.role); got != tc.want {
			t.Errorf("DefaultTierForRole(%q) = %q, want %q", tc.role, got, tc.want)
		}
	}
}

func TestNormalizeTier_AcceptsValidVariants(t *testing.T) {
	cases := []struct{ input, want string }{
		{"reasoning", "reasoning"},
		{"Reasoning", "reasoning"},
		{"REASONING", "reasoning"},
		{" fast ", "fast"},
		{"balanced", "balanced"},
	}
	for _, tc := range cases {
		got, ok := orchestration.NormalizeTier(tc.input)
		if !ok || string(got) != tc.want {
			t.Errorf("NormalizeTier(%q) = (%q, %v), want (%q, true)", tc.input, got, ok, tc.want)
		}
	}
}

func TestNormalizeTier_RejectsInvalid(t *testing.T) {
	for _, s := range []string{"genius", "", "  ", "fast fast"} {
		if _, ok := orchestration.NormalizeTier(s); ok {
			t.Errorf("NormalizeTier(%q) expected ok=false, got true", s)
		}
	}
}

func TestModelReference_ListsAllTiersInPlay(t *testing.T) {
	ref := orchestration.ModelReference([]orchestration.Tier{
		orchestration.TierReasoning, orchestration.TierBalanced, orchestration.TierFast,
	})
	for _, want := range []string{
		"claude-opus-4-7", "anthropic/claude-opus-4-7",
		"claude-sonnet-4-6", "anthropic/claude-sonnet-4-6",
		"claude-haiku-4-5-20251001", "anthropic/claude-haiku-4-5",
	} {
		if !strings.Contains(ref, want) {
			t.Errorf("ModelReference missing %q; got: %s", want, ref)
		}
	}
}

func TestModelReference_Deduplication(t *testing.T) {
	// Passing reasoning twice must produce ONE reasoning entry, not two.
	ref := orchestration.ModelReference([]orchestration.Tier{
		orchestration.TierReasoning, orchestration.TierReasoning,
	})
	if strings.Count(ref, "reasoning:") != 1 {
		t.Errorf("expected exactly one 'reasoning:' entry, got: %s", ref)
	}
	if strings.Contains(ref, "; ") {
		t.Errorf("expected single-entry (no semicolon separator), got: %s", ref)
	}
}

func TestModelReference_StableOrder(t *testing.T) {
	tiers := []orchestration.Tier{orchestration.TierFast, orchestration.TierReasoning}
	ref := orchestration.ModelReference(tiers)
	ri := strings.Index(ref, "reasoning:")
	fi := strings.Index(ref, "fast:")
	if ri == -1 || fi == -1 {
		t.Fatalf("expected both tiers in reference, got: %s", ref)
	}
	if ri > fi {
		t.Errorf("reasoning should appear before fast in stable order; got: %s", ref)
	}
}
