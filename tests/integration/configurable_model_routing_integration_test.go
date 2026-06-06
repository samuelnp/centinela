package integration_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// routingFromConfig mirrors cmd/centinela's orchestrationRouting: map the config
// leaf's accessors onto the domain resolver types.
func routingFromConfig(cfg *config.Config) (orchestration.RoleModels, orchestration.ModelMap) {
	models := orchestration.RoleModels{}
	for role, tier := range config.OrchestrationModelTiers(cfg) {
		models[role] = orchestration.RoleModel{Tier: tier}
	}
	for role, ov := range config.OrchestrationModelOverrides(cfg) {
		models[role] = orchestration.RoleModel{Overrides: ov}
	}
	return models, orchestration.ModelMap(config.OrchestrationModelMap(cfg))
}

func loadAt(t *testing.T, body string) *config.Config {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })       //nolint:errcheck
	os.Chdir(d)                                //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(body), 0644) //nolint:errcheck
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return cfg
}

// End-to-end config→resolver: tier remap from TOML resolves under opencode (AC1).
func TestRouting_ConfigToResolver_TierRemap(t *testing.T) {
	cfg := loadAt(t, "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n")
	models, mm := routingFromConfig(cfg)
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, models, mm, orchestration.RunnerOpenCode)
	if !ok || got != "moonshotai/kimi-k2" {
		t.Errorf("AC1 e2e: got (%q, %v)", got, ok)
	}
}

// End-to-end config→resolver: role override from TOML beats tier (AC2).
func TestRouting_ConfigToResolver_RoleOverride(t *testing.T) {
	cfg := loadAt(t, "[orchestration.models]\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n")
	models, mm := routingFromConfig(cfg)
	got, ok := orchestration.ResolveModel(orchestration.RoleSeniorEngineer, models, mm, orchestration.RunnerOpenCode)
	if !ok || got != "deepseek/deepseek-coder" {
		t.Errorf("AC2 e2e: got (%q, %v)", got, ok)
	}
}

// End-to-end: absent tables → built-in defaults (AC6).
func TestRouting_ConfigToResolver_AbsentDefaults(t *testing.T) {
	cfg := loadAt(t, "")
	models, mm := routingFromConfig(cfg)
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, models, mm, orchestration.RunnerClaude)
	if !ok || got != "claude-opus-4-7" {
		t.Errorf("AC6 e2e: got (%q, %v)", got, ok)
	}
}
