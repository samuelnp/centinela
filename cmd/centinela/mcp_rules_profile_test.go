package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestMCPRulesProfileAgreesWithEveryResolver: read_rules is the tool an agent
// consults to learn how it is governed, so its answer must equal what status,
// verdict and workflow_state resolve for the same tree. Reading
// cfg.Workflow.EnforcementProfile reported applyDefaults' normalized "strict"
// for every project whose centinela.toml merely omits the knob — which is what
// the scaffold template now ships — and "" for a project with no config at all.
func TestMCPRulesProfileAgreesWithEveryResolver(t *testing.T) {
	for _, tc := range []struct{ name, toml, want string }{
		{"no config at all", "", config.ProfileGuided},
		{"config without an explicit profile", "[gates]\nfile_size = true\n", config.ProfileGuided},
		{"explicit strict", "[workflow]\nenforcement_profile = \"strict\"\n", config.ProfileStrict},
		{"explicit outcome", "[workflow]\nenforcement_profile = \"outcome\"\n", config.ProfileOutcome},
		{"capable driver", "[orchestration]\ndriver_model = \"sonnet\"\n", config.ProfileGuided},
		{"limited driver", "[orchestration]\ndriver_model = \"haiku\"\n", config.ProfileStrict},
		{"unreadable config fails closed", "[workflow]\nuse_worktrees = tru\n", config.ProfileStrict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("CENTINELA_MODEL", "")
			if tc.toml != "" {
				os.WriteFile(config.Filename, []byte(tc.toml), 0644) //nolint:errcheck
			}
			if got := mcpRules().Profile; got != tc.want {
				t.Fatalf("read_rules profile = %q, want %q", got, tc.want)
			}
			// The same tree through the workflow-scoped resolver verdict uses.
			// The workflow deliberately carries NO pinned driver: a workflow that
			// predates the driver_model declaration is exactly the case that used
			// to make the reporting surfaces disagree with the acting ones.
			// Injecting DriverModelFrom here would resolve the driver FOR the
			// resolver and hide whether it consults the same source itself.
			cfg, _ := config.LoadForProfile()
			wf := &workflow.Workflow{ProfileContract: workflow.ProfileContractGuidedDefault}
			if got := workflow.EffectiveProfile(wf, cfg); got != tc.want {
				t.Fatalf("verdict resolver says %q, read_rules says %q", got, tc.want)
			}
			if got := rulesProfile(wf, cfg); got != tc.want {
				t.Fatalf("rulesProfile with an active workflow = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestMCPRulesStillReportsTheRestBestEffort: routing the profile through the
// shared resolver must not have turned read_rules into a failing tool — the
// other fields stay best-effort, and an unreadable config simply omits them.
func TestMCPRulesStillReportsTheRestBestEffort(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(config.Filename,
		[]byte("[gates]\nfile_size = true\n[i18n]\nlocales = [\"en\"]\n"), 0644) //nolint:errcheck
	out := mcpRules()
	if out.MaxFileLines != 100 || len(out.Locales) != 1 || len(out.Gates) == 0 {
		t.Fatalf("read_rules must still report the rule surface: %+v", out)
	}

	t.Chdir(t.TempDir())
	os.WriteFile(config.Filename, []byte("[workflow]\nuse_worktrees = tru\n"), 0644) //nolint:errcheck
	broken := mcpRules()
	if broken.Profile != config.ProfileStrict {
		t.Fatalf("a broken config must report strict, got %q", broken.Profile)
	}
	if len(broken.Locales) != 0 || len(broken.Gates) != 0 {
		t.Fatalf("a broken config must not invent a rule surface: %+v", broken)
	}
}
