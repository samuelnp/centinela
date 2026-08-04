package config

import (
	"os"
	"strings"
	"testing"
)

// TestLoadForProfileFailsClosed: an unreadable centinela.toml makes the profile
// unknowable, so the returned config must resolve to strict on EVERY tier that
// reads it — and it must never be nil, because a nil config is precisely what
// falls through to the shipped guided tail.
func TestLoadForProfileFailsClosed(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\nuse_worktrees = tru\n"), 0644) //nolint:errcheck
	cfg, err := LoadForProfile()
	if err == nil {
		t.Fatal("the real load error must still be returned, not swallowed")
	}
	if cfg == nil {
		t.Fatal("LoadForProfile must never return a nil config")
	}
	// The fallback marks itself unreadable rather than faking an explicit pin:
	// faking one made every surface attribute the profile to a global setting the
	// operator never wrote. Fail-closed direction is unchanged.
	if !cfg.LoadFailed() {
		t.Fatal("the fallback must mark itself as an unreadable config")
	}
	if cfg.Workflow.RawEnforcementProfile != "" {
		t.Fatalf("the fallback must not fake an explicit pin, got %q",
			cfg.Workflow.RawEnforcementProfile)
	}
	if got := ProjectDefaultProfile(cfg); got != ProfileStrict {
		t.Fatalf("ProjectDefaultProfile on the fallback = %q, want strict", got)
	}
	if !ProfileDefaults(ProjectDefaultProfile(cfg)).RequireRoadmapGrading {
		t.Fatal("the fallback must keep the full greenfield cascade")
	}
}

// TestLoadForProfileFailsClosedEvenOverADeclaredDriver: the fallback pins the
// global tier, which outranks the capability tier — so a broken config that
// declares a frontier model still reports strict rather than outcome.
func TestLoadForProfileFailsClosedEvenOverADeclaredDriver(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename,
		[]byte("[orchestration]\ndriver_model = \"opus\"\n[workflow]\nuse_worktrees = tru\n"), 0644) //nolint:errcheck
	cfg, err := LoadForProfile()
	if err == nil || !strings.Contains(err.Error(), Filename) {
		t.Fatalf("the error must name the file that failed, got %v", err)
	}
	if got := ProjectDefaultProfile(cfg); got != ProfileStrict {
		t.Fatalf("a broken config must not inherit a driver's profile, got %q", got)
	}
}

// TestLoadForProfilePassesThroughAGoodConfig is the ✅ direction, both ways: a
// readable config is returned untouched, and an ABSENT centinela.toml is not a
// read failure — it is the zero-config case that legitimately takes guided.
func TestLoadForProfilePassesThroughAGoodConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("CENTINELA_MODEL", "")
	cfg, err := LoadForProfile()
	if err != nil {
		t.Fatalf("an absent config is not a failure: %v", err)
	}
	if got := ProjectDefaultProfile(cfg); got != ProfileGuided {
		t.Fatalf("zero config must still take the guided default, got %q", got)
	}
	os.WriteFile(Filename, []byte("[workflow]\nenforcement_profile = \"outcome\"\n"), 0644) //nolint:errcheck
	cfg, err = LoadForProfile()
	if err != nil {
		t.Fatalf("LoadForProfile: %v", err)
	}
	if got := ProjectDefaultProfile(cfg); got != ProfileOutcome {
		t.Fatalf("an explicit profile must survive the helper, got %q", got)
	}
}
