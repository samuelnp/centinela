package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// cfgWithDriver writes a real centinela.toml declaring driver_model and loads
// it. Nothing is injected: the resolver has to find the driver the same way the
// acting paths do, which is the only way this test can see the bug it guards.
func cfgWithDriver(t *testing.T, toml string) *config.Config {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("CENTINELA_MODEL", "")
	if toml != "" {
		os.WriteFile(config.Filename, []byte(toml), 0644) //nolint:errcheck
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

// TestDeclaredDriverAppliesToWorkflowsWithoutThePin: a workflow that predates
// the driver_model declaration carries no pin, and the capability tier must
// still engage for it. Reading only wf.DriverModel made every REPORTING surface
// answer guided while every ACTING surface behaved strict on the same tree.
func TestDeclaredDriverAppliesToWorkflowsWithoutThePin(t *testing.T) {
	for _, tc := range []struct{ name, toml, want string }{
		{"limited driver", "[orchestration]\ndriver_model = \"haiku\"\n", config.ProfileStrict},
		{"capable driver", "[orchestration]\ndriver_model = \"sonnet\"\n", config.ProfileGuided},
		{"frontier driver", "[orchestration]\ndriver_model = \"opus\"\n", config.ProfileOutcome},
		{"unmapped driver", "[orchestration]\ndriver_model = \"who/knows\"\n", config.ProfileStrict},
		{"no driver declared", "", config.ProfileGuided},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := cfgWithDriver(t, tc.toml)
			wf := &Workflow{ProfileContract: ProfileContractGuidedDefault} // no pin
			if got := EffectiveProfile(wf, cfg); got != tc.want {
				t.Errorf("EffectiveProfile = %q, want %q", got, tc.want)
			}
			reported, note := ProfileProvenance(wf, cfg)
			if reported != tc.want {
				t.Errorf("ProfileProvenance = %q, want %q (note %q)", reported, tc.want, note)
			}
			// The acting paths resolve the same tree through ProjectDefaultProfile.
			if acting := config.ProjectDefaultProfile(cfg); acting != tc.want {
				t.Errorf("acting paths resolve %q, reporting resolves %q", acting, tc.want)
			}
		})
	}
}

// TestEnvDriverModelReachesTheResolvers: $CENTINELA_MODEL is one of the sources
// DriverModelFrom consults, so the reporting surfaces must see it too.
func TestEnvDriverModelReachesTheResolvers(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wf := &Workflow{ProfileContract: ProfileContractGuidedDefault}
	if got := EffectiveProfile(wf, cfg); got != config.ProfileGuided {
		t.Fatalf("without an env model the tail applies, got %q", got)
	}
	t.Setenv("CENTINELA_MODEL", "haiku")
	if got := EffectiveProfile(wf, cfg); got != config.ProfileStrict {
		t.Fatalf("an env-declared limited driver must resolve strict, got %q", got)
	}
	if got := config.ProjectDefaultProfile(cfg); got != config.ProfileStrict {
		t.Fatalf("the acting paths must agree, got %q", got)
	}
}

// TestWorkflowPinOutranksTheProjectDeclaration: the pin is the state-dated
// decision for that workflow and must survive a later project-level change.
func TestWorkflowPinOutranksTheProjectDeclaration(t *testing.T) {
	cfg := cfgWithDriver(t, "[orchestration]\ndriver_model = \"haiku\"\n")
	pinnedToCapable := &Workflow{
		ProfileContract: ProfileContractGuidedDefault,
		DriverModel:     "sonnet",
	}
	if got := EffectiveProfile(pinnedToCapable, cfg); got != config.ProfileGuided {
		t.Fatalf("the workflow's own pin must win, got %q", got)
	}
	if got := TierDriverModel(pinnedToCapable, cfg); got != "sonnet" {
		t.Fatalf("TierDriverModel = %q, want the pinned model", got)
	}
	if got := TierDriverModel(&Workflow{}, cfg); got != "haiku" {
		t.Fatalf("without a pin it must fall through to the declaration, got %q", got)
	}
}
