package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// routingFromConfig (shared with configurable_model_routing_integration_test.go)
// reconstructs the config-leaf → domain mapping that
// cmd/centinela/hook_orchestration.go's orchestrationRouting performs.

// annotateRolesHelper mirrors cmd/centinela/orchestration_annotate.go: it
// resolves every runner's concrete model via the real domain resolver and
// emits the per-runner annotation, so this test exercises production logic.
func annotateRolesHelper(roles []orchestration.Role, models orchestration.RoleModels, modelMap orchestration.ModelMap) (names []string, tiers []orchestration.Tier) {
	for _, role := range roles {
		parts := make([]string, 0, len(orchestration.AllowedRunnerKeys()))
		for _, key := range orchestration.AllowedRunnerKeys() {
			id, _ := orchestration.ResolveModel(role, models, modelMap, orchestration.Runner(key))
			parts = append(parts, "model: "+id+" ("+key+")")
		}
		names = append(names, string(role)+" ("+strings.Join(parts, ", ")+")")
		tiers = append(tiers, orchestration.RoleTier(role, models))
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
	models, modelMap := routingFromConfig(cfg)
	roles := []orchestration.Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial}
	names, tiers := annotateRolesHelper(roles, models, modelMap)
	if !strings.Contains(names[0], "big-thinker (model: claude-haiku-4-5-20251001 (claude)") {
		t.Errorf("expected big-thinker annotated with fast→haiku, got %q", names[0])
	}
	if !strings.Contains(names[1], "feature-specialist (model: claude-sonnet-4-6 (claude)") {
		t.Errorf("expected feature-specialist annotated with balanced default, got %q", names[1])
	}
	if ref := orchestration.ModelReference(tiers); !strings.Contains(ref, "fast:") {
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
	models, modelMap := routingFromConfig(cfg)
	roles := orchestration.RequiredRoles("plan")
	names, _ := annotateRolesHelper(roles, models, modelMap)
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
