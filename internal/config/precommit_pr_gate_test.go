package config

import (
	"os"
	"testing"
)

func TestNormalizePrecommit_OmittedDefaultsSkipBuildTrue(t *testing.T) {
	got := NormalizePrecommit(PrecommitConfig{Enabled: true})
	if !got.SkipBuild {
		t.Fatalf("omitted skip_build (RawSkipBuild nil) must default to true, got %+v", got)
	}
	if !got.Enabled {
		t.Fatalf("Enabled must round-trip, got %+v", got)
	}
}

func TestNormalizePrecommit_ExplicitFalseIsHonored(t *testing.T) {
	f := false
	got := NormalizePrecommit(PrecommitConfig{RawSkipBuild: &f})
	if got.SkipBuild {
		t.Fatalf("explicit skip_build=false must be honored, got %+v", got)
	}
	tr := true
	if got := NormalizePrecommit(PrecommitConfig{RawSkipBuild: &tr}); !got.SkipBuild {
		t.Fatalf("explicit skip_build=true must stay true, got %+v", got)
	}
}

func TestValidatePrecommitAndPrGate_NoOp(t *testing.T) {
	if err := validatePrecommit(PrecommitConfig{}); err != nil {
		t.Fatalf("validatePrecommit must be a no-op, got %v", err)
	}
	if err := validatePrGate(PrGateConfig{}); err != nil {
		t.Fatalf("validatePrGate must be a no-op, got %v", err)
	}
}

func TestNormalizePrGate_Defaults(t *testing.T) {
	got := NormalizePrGate(PrGateConfig{Enabled: true})
	if got.FailOnWarning {
		t.Fatalf("fail_on_warning must default to false, got %+v", got)
	}
	if !got.Enabled {
		t.Fatalf("Enabled must round-trip, got %+v", got)
	}
}

// The whole point of RawSkipBuild: an explicit `skip_build = false` decoded from
// TOML through Load must remain false after applyDefaults, distinct from an
// omitted section which defaults to true.
func TestLoad_PrecommitSkipBuildExplicitFalse(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[precommit]\nskip_build = false\n"), 0o644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Precommit.SkipBuild {
		t.Fatalf("explicit skip_build=false must survive Load, got %+v", cfg.Precommit)
	}
}

func TestLoad_PrecommitSkipBuildOmittedDefaultsTrue(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[precommit]\nenabled = true\n"), 0o644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !cfg.Precommit.SkipBuild {
		t.Fatalf("omitted skip_build must default to true through Load, got %+v", cfg.Precommit)
	}
}
