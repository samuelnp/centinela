package config

import "strings"

// DiffMode values accepted in [validate] diff_mode.
const (
	DiffModeAuto   = "auto"
	DiffModeAlways = "always"
	DiffModeOff    = "off"
)

// Mode is the resolved validate execution mode after applying defaults,
// CI detection, and CLI flag overrides.
type Mode int

const (
	ModeFull Mode = iota
	ModeChanged
)

// Env carries the environment signals that influence mode resolution.
// Kept as an interface-shaped struct so tests can stub it deterministically.
type Env struct {
	CI bool
}

// IsCI returns whether the resolver should treat this run as CI.
func (e Env) IsCI() bool { return e.CI }

// FlagOverride represents the result of `--changed` / `--full` CLI flags.
type FlagOverride int

const (
	FlagNone FlagOverride = iota
	FlagForceChanged
	FlagForceFull
)

// NormalizeDiffMode coerces the configured mode to a known value,
// defaulting to "auto" when empty or unrecognized.
func NormalizeDiffMode(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case DiffModeAlways:
		return DiffModeAlways
	case DiffModeOff:
		return DiffModeOff
	default:
		return DiffModeAuto
	}
}

// NormalizeDiffBase returns the configured base or "main" when empty.
func NormalizeDiffBase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "main"
	}
	return s
}

// ResolveMode applies the precedence: CLI flag > config mode > CI env.
func (v ValidateConfig) ResolveMode(env Env, flag FlagOverride) Mode {
	switch flag {
	case FlagForceChanged:
		return ModeChanged
	case FlagForceFull:
		return ModeFull
	}
	switch NormalizeDiffMode(v.DiffMode) {
	case DiffModeAlways:
		return ModeChanged
	case DiffModeOff:
		return ModeFull
	default:
		if env.IsCI() {
			return ModeFull
		}
		return ModeChanged
	}
}
