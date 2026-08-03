package doctor

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// cfgPinned returns a config whose enforcement_profile was set explicitly.
func cfgPinned(profile string) *config.Config {
	c := &config.Config{}
	c.Workflow.EnforcementProfile = profile
	c.Workflow.RawEnforcementProfile = profile
	return c
}

// TestProfileDefaultAdvisesInheritedDefault is the ✅ direction: workflows
// exist, nothing is pinned, so the project is told what it inherited.
func TestProfileDefaultAdvisesInheritedDefault(t *testing.T) {
	repoFixture(t)
	seedWorkflow(t, "feat")
	d := profileDefaultCheck{}.Run(Context{Root: ".", Config: &config.Config{}})
	if d.Status != Warn {
		t.Fatalf("status = %v, want Warn (advisory)", d.Status)
	}
	if !strings.Contains(d.Message, config.ProfileGuided) {
		t.Fatalf("message must name guided, got %q", d.Message)
	}
	if !strings.Contains(strings.Join(d.Details, "\n"), `enforcement_profile = "strict"`) {
		t.Fatalf("details must show the exact pin line, got %v", d.Details)
	}
}

// TestProfileDefaultSilentWhenPinnedOrEmpty is the ❌ direction, both ways.
func TestProfileDefaultSilentWhenPinnedOrEmpty(t *testing.T) {
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		repoFixture(t)
		seedWorkflow(t, "feat")
		if d := (profileDefaultCheck{}).Run(Context{Root: ".", Config: cfgPinned(profile)}); d.Status != OK {
			t.Fatalf("an explicit %q must silence the advisory, got %v", profile, d.Status)
		}
	}
	repoFixture(t) // workflow-less project: nothing has inherited anything yet
	if d := (profileDefaultCheck{}).Run(Context{Root: ".", Config: &config.Config{}}); d.Status != OK {
		t.Fatalf("a project with no workflows must be silent, got %v", d.Status)
	}
}

// TestProfileDefaultIsNeverFatalAndNeverFixed: the advisory must not change the
// doctor exit code, and --fix must not rewrite the user's config.
func TestProfileDefaultIsNeverFatalAndNeverFixed(t *testing.T) {
	repoFixture(t)
	seedWorkflow(t, "feat")
	d := profileDefaultCheck{}.Run(Context{Root: ".", Config: &config.Config{}})
	if ExitError([]Diagnosis{d}) {
		t.Fatal("the profile-default advisory must never fail the doctor run")
	}
	if d.Repair != nil {
		t.Fatal("the advisory must carry no repair — no config is ever rewritten")
	}
	if d.Name != "profile-default" {
		t.Fatalf("name = %q, want profile-default", d.Name)
	}
}

// TestProfileDefaultRegistered: an unregistered check diagnoses nothing.
func TestProfileDefaultRegistered(t *testing.T) {
	for _, c := range checks() {
		if c.Name() == "profile-default" {
			return
		}
	}
	t.Fatal("profileDefaultCheck must be in the doctor registry")
}
