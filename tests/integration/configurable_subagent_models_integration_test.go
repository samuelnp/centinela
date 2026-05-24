package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// annotateRolesHelper calls DefaultTierForRole + NormalizeTier exactly as the
// cmd/centinela/orchestration_annotate.go helper does, in-process.
func annotateRolesHelper(roles []orchestration.Role, models map[string]string) (names []string, tiers []orchestration.Tier) {
	for _, role := range roles {
		tier := orchestration.DefaultTierForRole(role)
		if raw, ok := models[string(role)]; ok {
			if norm, valid := orchestration.NormalizeTier(raw); valid {
				tier = norm
			}
		}
		names = append(names, string(role)+" (model: "+string(tier)+")")
		tiers = append(tiers, tier)
	}
	return names, tiers
}

func TestAnnotateRoles_WithConfigOverride(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                                             //nolint:errcheck
	os.Chdir(d)                                                                                      //nolint:errcheck
	os.WriteFile("centinela.toml", []byte("[orchestration.models]\nbig-thinker = \"fast\"\n"), 0644) //nolint:errcheck
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	models := config.OrchestrationModels(cfg)
	roles := []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial}
	names, tiers := annotateRolesHelper(roles, models)
	if !strings.Contains(names[0], "big-thinker (model: fast)") {
		t.Errorf("expected big-thinker annotated with fast, got %q", names[0])
	}
	if !strings.Contains(names[1], "feature-specialist (model: balanced)") {
		t.Errorf("expected feature-specialist annotated with balanced (default), got %q", names[1])
	}
	ref := orchestration.ModelReference(tiers)
	if !strings.Contains(ref, "fast:") {
		t.Errorf("model reference should include fast tier, got: %s", ref)
	}
}

func TestAnnotateRoles_NoConfig_AllDefaults(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load without toml: %v", err)
	}
	models := config.OrchestrationModels(cfg)
	roles := orchestration.RequiredRoles("plan")
	names, _ := annotateRolesHelper(roles, models)
	for _, name := range names {
		if !strings.Contains(name, "(model: ") {
			t.Errorf("expected every role annotated, got: %q", name)
		}
	}
}

func TestAnnotateRoles_ModelReferenceLineContainsBothRunners(t *testing.T) {
	tiers := []orchestration.Tier{orchestration.TierReasoning}
	ref := orchestration.ModelReference(tiers)
	if !strings.Contains(ref, "claude-opus-4-7") {
		t.Errorf("reference missing claude ID: %s", ref)
	}
	if !strings.Contains(ref, "anthropic/claude-opus-4-7") {
		t.Errorf("reference missing opencode ID: %s", ref)
	}
}

func TestOrchestrationModelsAccessor_NilSafe(t *testing.T) {
	if m := config.OrchestrationModels(nil); m != nil {
		t.Errorf("expected nil for nil config, got %v", m)
	}
}
