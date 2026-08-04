package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestProfileProvenance_UnresolvedConfigIsNamed: a pinned workflow under a
// config nobody loaded resolves strict, and the NOTE says why. Reporting
// "default (strict, legacy workflow)" there would be a lie — the workflow is not
// legacy — and the operator needs to know which condition they are in.
func TestProfileProvenance_UnresolvedConfigIsNamed(t *testing.T) {
	pinnedWf := &Workflow{ProfileContract: ProfileContractGuidedDefault}
	profile, note := ProfileProvenance(pinnedWf, &config.Config{})
	if profile != config.ProfileStrict || note != "default (strict, unresolved config)" {
		t.Fatalf("got (%q,%q), want strict + the unresolved-config note", profile, note)
	}
	if _, legacyNote := ProfileProvenance(&Workflow{}, &config.Config{}); legacyNote == note {
		t.Fatal("an unreadable config and a legacy workflow must stay distinguishable")
	}
}

// TestProfileProvenance_UnreadableConfigIsAttributedHonestly: an unreadable
// centinela.toml must not be reported as a "global" pin. Faking an explicit pin
// on the fallback made status say `Profile strict (global)` for a chmod-000 file
// — attributing the profile to a setting the operator may never have written —
// and left the honest branch unreachable.
func TestProfileProvenance_UnreadableConfigIsAttributedHonestly(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(config.Filename, []byte("[workflow\nenforcement_profile = \"guided\"\n"), 0644) //nolint:errcheck
	cfg, err := config.LoadForProfile()
	if err == nil {
		t.Fatal("fixture check: this config must fail to parse")
	}
	for _, wf := range []*Workflow{
		{ProfileContract: ProfileContractGuidedDefault},
		{}, // legacy
		{DriverModel: "opus", ProfileContract: ProfileContractGuidedDefault},
	} {
		profile, note := ProfileProvenance(wf, cfg)
		if profile != config.ProfileStrict {
			t.Fatalf("an unreadable config must resolve strict, got %q", profile)
		}
		if note != "unreadable "+config.Filename+", default strict" {
			t.Fatalf("note = %q, want the unreadable-config attribution", note)
		}
		if got := EffectiveProfile(wf, cfg); got != profile {
			t.Fatalf("EffectiveProfile %q disagrees with provenance %q", got, profile)
		}
	}
	// Even a declared frontier model cannot lift an unreadable config to outcome.
	if got := config.ProjectDefaultProfile(cfg); got != config.ProfileStrict {
		t.Fatalf("the acting paths must agree, got %q", got)
	}
}
