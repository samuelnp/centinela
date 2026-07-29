package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyPlanModelRoleNotice_NamesKeysAndPlanner(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{
		"big-thinker":        {Tier: "reasoning"},
		"feature-specialist": {Tier: "balanced"},
	})
	notice := LegacyPlanModelRoleNotice(cfg)
	for _, want := range []string{"big-thinker", "feature-specialist", "planner"} {
		if !strings.Contains(notice, want) {
			t.Fatalf("notice %q missing %q", notice, want)
		}
	}
}

func TestLegacyPlanModelRoleNotice_EmptyWithoutLegacyKeys(t *testing.T) {
	if LegacyPlanModelRoleNotice(nil) != "" {
		t.Fatal("nil config must render no notice")
	}
	cfg := modelsCfg(map[string]RoleModelValue{"planner": {Tier: "reasoning"}})
	if got := LegacyPlanModelRoleNotice(cfg); got != "" {
		t.Fatalf("a migrated config must render no notice, got %q", got)
	}
}

func TestLegacyPlanModelRoleNotice_SingleKeyHasNoTrailingSeparator(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{"big-thinker": {Tier: "fast"}})
	notice := LegacyPlanModelRoleNotice(cfg)
	if strings.Contains(notice, "big-thinker,") {
		t.Fatalf("single key must not carry a separator: %q", notice)
	}
}

// D8's non-negotiable: a centinela.toml carrying only legacy plan keys must
// still LOAD. Removing them from the allow-list would brick existing projects.
func TestLoad_LegacyPlanModelKeysStillAccepted(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	body := "[orchestration.models]\nbig-thinker = \"reasoning\"\nfeature-specialist = \"balanced\"\n"
	if err := os.WriteFile(filepath.Join(dir, Filename), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("legacy-key config must still load: %v", err)
	}
	if got := OrchestrationModelTiers(cfg)["planner"]; got != "reasoning" {
		t.Fatalf("legacy key must resolve for planner, got %q", got)
	}
	if LegacyPlanModelRoleNotice(cfg) == "" {
		t.Fatal("a legacy-key config must surface the deprecation notice")
	}
}

func TestLoad_PlannerModelKeyAccepted(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	body := "[orchestration.models]\nplanner = \"reasoning\"\n"
	if err := os.WriteFile(filepath.Join(dir, Filename), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("planner key must be accepted: %v", err)
	}
	if got := OrchestrationModelTiers(cfg)["planner"]; got != "reasoning" {
		t.Fatalf("planner tier = %q", got)
	}
}
