package orchestration

import (
	"strings"
	"testing"
)

func TestDefaultTierForRole_Defaults(t *testing.T) {
	cases := []struct {
		role Role
		want Tier
	}{
		{RoleBigThinker, TierReasoning},
		{RoleSeniorEngineer, TierReasoning},
		{RoleFeatureSpecial, TierBalanced},
		{RoleQASeniorEngineer, TierBalanced},
		{RoleUXUISpecialist, TierBalanced},
		{RoleDocsSpecialist, TierFast},
		{RoleValidationSpec, TierFast},
		{RoleMergeSteward, TierReasoning},
		{Role("gatekeeper"), TierFast},
		{Role("edge-case-tester"), TierFast},
	}
	for _, tc := range cases {
		if got := DefaultTierForRole(tc.role); got != tc.want {
			t.Errorf("DefaultTierForRole(%q) = %q, want %q", tc.role, got, tc.want)
		}
	}
}

func TestDefaultTierForRole_UnknownRoleReturnsBalanced(t *testing.T) {
	if got := DefaultTierForRole(Role("unknown-role")); got != TierBalanced {
		t.Errorf("expected TierBalanced for unknown role, got %q", got)
	}
}

func TestNormalizeTier_Valid(t *testing.T) {
	cases := []struct {
		in   string
		want Tier
	}{
		{"reasoning", TierReasoning},
		{"Reasoning", TierReasoning},
		{" fast ", TierFast},
		{"BALANCED", TierBalanced},
	}
	for _, tc := range cases {
		got, ok := NormalizeTier(tc.in)
		if !ok || got != tc.want {
			t.Errorf("NormalizeTier(%q) = (%q, %v), want (%q, true)", tc.in, got, ok, tc.want)
		}
	}
}

func TestNormalizeTier_Invalid(t *testing.T) {
	for _, s := range []string{"genius", "", "  ", "fast fast"} {
		if _, ok := NormalizeTier(s); ok {
			t.Errorf("NormalizeTier(%q): expected ok=false", s)
		}
	}
}

func TestAllowedTiers_ContainsThree(t *testing.T) {
	tiers := AllowedTiers()
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d: %v", len(tiers), tiers)
	}
}

func TestAllowedRoleSlugs_NotEmpty(t *testing.T) {
	slugs := AllowedRoleSlugs()
	if len(slugs) == 0 {
		t.Fatal("AllowedRoleSlugs returned empty slice")
	}
	found := false
	for _, s := range slugs {
		if s == string(RoleBigThinker) {
			found = true
		}
		if strings.TrimSpace(s) != s {
			t.Errorf("AllowedRoleSlugs entry %q has surrounding whitespace", s)
		}
	}
	if !found {
		t.Errorf("AllowedRoleSlugs missing %q", RoleBigThinker)
	}
}
