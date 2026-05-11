package config

import "testing"

func TestNormalizeDiffMode(t *testing.T) {
	cases := map[string]string{
		"":         DiffModeAuto,
		"  ":       DiffModeAuto,
		"auto":     DiffModeAuto,
		"Always":   DiffModeAlways,
		"  off  ":  DiffModeOff,
		"bogus":    DiffModeAuto,
		"OFF":      DiffModeOff,
		"ALWAYS":   DiffModeAlways,
	}
	for in, want := range cases {
		if got := NormalizeDiffMode(in); got != want {
			t.Fatalf("NormalizeDiffMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeDiffBase(t *testing.T) {
	if got := NormalizeDiffBase(""); got != "main" {
		t.Fatalf("empty should default to main, got %q", got)
	}
	if got := NormalizeDiffBase("  develop  "); got != "develop" {
		t.Fatalf("expected trimmed develop, got %q", got)
	}
	if got := NormalizeDiffBase("master"); got != "master" {
		t.Fatalf("expected master preserved, got %q", got)
	}
}

func TestResolveModeFlagOverrides(t *testing.T) {
	v := ValidateConfig{DiffMode: DiffModeOff}
	if got := v.ResolveMode(Env{CI: true}, FlagForceChanged); got != ModeChanged {
		t.Fatalf("--changed must force Changed regardless of mode/CI")
	}
	v = ValidateConfig{DiffMode: DiffModeAlways}
	if got := v.ResolveMode(Env{CI: false}, FlagForceFull); got != ModeFull {
		t.Fatalf("--full must force Full regardless of mode/CI")
	}
}

func TestResolveModeAlwaysAndOff(t *testing.T) {
	always := ValidateConfig{DiffMode: DiffModeAlways}
	if always.ResolveMode(Env{CI: true}, FlagNone) != ModeChanged {
		t.Fatalf("always must stay Changed in CI")
	}
	if always.ResolveMode(Env{CI: false}, FlagNone) != ModeChanged {
		t.Fatalf("always must stay Changed locally")
	}
	off := ValidateConfig{DiffMode: DiffModeOff}
	if off.ResolveMode(Env{CI: true}, FlagNone) != ModeFull {
		t.Fatalf("off must stay Full in CI")
	}
	if off.ResolveMode(Env{CI: false}, FlagNone) != ModeFull {
		t.Fatalf("off must stay Full locally")
	}
}

func TestResolveModeAutoFlipsOnCI(t *testing.T) {
	auto := ValidateConfig{DiffMode: DiffModeAuto}
	if auto.ResolveMode(Env{CI: false}, FlagNone) != ModeChanged {
		t.Fatalf("auto must be Changed locally")
	}
	if auto.ResolveMode(Env{CI: true}, FlagNone) != ModeFull {
		t.Fatalf("auto must be Full in CI")
	}
	empty := ValidateConfig{}
	if empty.ResolveMode(Env{CI: false}, FlagNone) != ModeChanged {
		t.Fatalf("empty diff_mode must default to auto/local Changed")
	}
}

func TestEnvIsCI(t *testing.T) {
	if (Env{CI: true}).IsCI() != true {
		t.Fatalf("IsCI must mirror CI field")
	}
	if (Env{}).IsCI() != false {
		t.Fatalf("zero Env must not be CI")
	}
}
