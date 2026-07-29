package doctor

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func planKeyConfig(keys ...string) *config.Config {
	cfg := &config.Config{}
	cfg.Verify.TimeoutSeconds = 240
	cfg.Orchestration.Models = map[string]config.RoleModelValue{}
	for _, k := range keys {
		cfg.Orchestration.Models[k] = config.RoleModelValue{Tier: "reasoning"}
	}
	return cfg
}

// D8: doctor surfaces the retired plan-role keys and suggests planner.
func TestConfigCheck_ReportsRetiredPlanRoleKeys(t *testing.T) {
	repoFixture(t)
	d := configCheck{}.Run(Context{Config: planKeyConfig("big-thinker", "feature-specialist")})
	if d.Status != Warn {
		t.Fatalf("retired plan keys must Warn (advisory), got %v", d.Status)
	}
	joined := strings.Join(d.Details, " ")
	for _, want := range []string{"big-thinker", "feature-specialist", "planner"} {
		if !strings.Contains(joined, want) {
			t.Errorf("details must mention %q: %v", want, d.Details)
		}
	}
}

// A migrated config stays clean — the notice must not be permanent noise.
func TestConfigCheck_NoNoticeAfterMigration(t *testing.T) {
	repoFixture(t)
	d := configCheck{}.Run(Context{Config: planKeyConfig("planner")})
	if d.Status != OK {
		t.Fatalf("a migrated config must be OK, got %v %q %v", d.Status, d.Message, d.Details)
	}
}
