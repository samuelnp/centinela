package doctor

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestProfileDefaultSaysNothingWhenTheConfigIsUnreadable: a parse error leaves
// ctx.Config nil, which is NOT the same as "no pin". Reading it as unpinned told
// an operator whose config pins strict that they were inheriting guided — the
// doctor, of all surfaces, answering with the loosest profile. The config check
// already reports the parse failure as an ERROR, so silence here loses nothing.
func TestProfileDefaultSaysNothingWhenTheConfigIsUnreadable(t *testing.T) {
	repoFixture(t)
	seedWorkflow(t, "feat")
	for _, ctx := range []Context{
		{Root: ".", Config: nil, CfgErr: errStub},
		{Root: ".", Config: cfgPinned(config.ProfileStrict), CfgErr: errStub},
		{Root: ".", Config: nil},
	} {
		d := profileDefaultCheck{}.Run(ctx)
		if d.Status != OK {
			t.Fatalf("an unreadable config must not raise the inherited-guided advisory: %v", d)
		}
		if strings.Contains(d.Message, config.ProfileGuided) {
			t.Fatalf("the message must not claim guided when the profile is unknown: %q", d.Message)
		}
		if ExitError([]Diagnosis{d}) {
			t.Fatal("this check must never fail the doctor run")
		}
	}
}

// TestProfileDefaultSilentWhenTheProjectDoesNotActuallyInheritGuided: a declared
// driver model resolves through the capability tier, so an unmapped one lands on
// strict even with no explicit pin. Advising "you are on guided" there repeats,
// at project scope, the mistake this check was hardened against at parse scope.
func TestProfileDefaultSilentWhenTheProjectDoesNotActuallyInheritGuided(t *testing.T) {
	for _, toml := range []string{
		"[orchestration]\ndriver_model = \"who/knows\"\n", // unmapped → strict
		"[orchestration]\ndriver_model = \"haiku\"\n",     // limited → strict
		"[orchestration]\ndriver_model = \"opus\"\n",      // frontier → outcome
	} {
		repoFixture(t)
		seedWorkflow(t, "feat")
		writeFile(t, config.Filename, toml)
		t.Setenv("CENTINELA_MODEL", "")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if d := (profileDefaultCheck{}).Run(Context{Root: ".", Config: cfg}); d.Status != OK {
			t.Fatalf("project resolving %q must draw no guided advisory: %v",
				config.ProjectDefaultProfile(cfg), d)
		}
	}
}
