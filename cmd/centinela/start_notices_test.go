package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func planKeysCfg(keys ...string) *config.Config {
	cfg := &config.Config{}
	cfg.Orchestration.Models = map[string]config.RoleModelValue{}
	for _, k := range keys {
		cfg.Orchestration.Models[k] = config.RoleModelValue{Tier: "reasoning"}
	}
	return cfg
}

func TestPrintLegacyPlanRoleNotice_PrintsOnceForRetiredKeys(t *testing.T) {
	out := captureStdout(t, func() {
		printLegacyPlanRoleNotice(planKeysCfg("big-thinker"))
	})
	if !strings.Contains(out, "big-thinker") || !strings.Contains(out, "planner") {
		t.Fatalf("start must surface the migration notice, got %q", out)
	}
	if strings.Count(out, "planner") == 0 {
		t.Fatalf("notice must name planner: %q", out)
	}
}

func TestPrintLegacyPlanRoleNotice_SilentWhenMigrated(t *testing.T) {
	out := captureStdout(t, func() {
		printLegacyPlanRoleNotice(planKeysCfg("planner"))
	})
	if strings.TrimSpace(out) != "" {
		t.Fatalf("a migrated config must print nothing, got %q", out)
	}
	out = captureStdout(t, func() { printLegacyPlanRoleNotice(nil) })
	if strings.TrimSpace(out) != "" {
		t.Fatalf("nil config must print nothing, got %q", out)
	}
}
