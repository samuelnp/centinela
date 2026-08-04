package config

// LoadForProfile loads centinela.toml for any surface that RESOLVES or REPORTS
// an enforcement profile. It is the single entry point for that decision, and it
// fails CLOSED: it never returns nil, and on a read or parse failure it returns a
// config that pins strict EXPLICITLY, alongside the error.
//
// The failure direction is the entire point. A nil config makes every resolver
// skip the explicit-global and capability tiers and land on the shipped guided
// tail, so one typo in a centinela.toml that says enforcement_profile = "strict"
// would make the reporting surfaces answer with the LOOSEST profile while the
// enforcing surfaces refuse outright. Fixing that one call site at a time is
// whack-a-mole; every profile surface routes through here instead, so a new
// surface inherits the direction rather than choosing it.
//
// Callers that can surface the error should: the returned error is the real one
// from Load, never swallowed here. Callers that cannot (status rendering, the
// MCP read_rules tool) may discard it and still be correct, because the config
// they get back already answers strict.
func LoadForProfile() (*Config, error) {
	cfg, err := Load()
	if err == nil {
		return cfg, nil
	}
	return strictFallbackConfig(), err
}

// strictFallbackConfig is the stand-in used when the real config is unreadable.
// It sets loadFailed rather than faking an explicit enforcement_profile pin.
//
// Faking the pin worked, but it lied: every surface then attributed the profile
// to a global setting the operator may never have written ("Profile strict
// (global)" on a chmod-000 file), and it made the honest "unreadable config"
// provenance unreachable. loadFailed is checked ahead of the capability and tail
// tiers by all three resolvers, so the fail-closed direction is unchanged — a
// declared frontier model cannot lift an unreadable config to outcome.
func strictFallbackConfig() *Config {
	cfg := &Config{}
	cfg.loadFailed = true
	return cfg
}

// LoadFailed reports whether this Config is the stand-in returned when
// centinela.toml could not be read. Nil-safe. Resolvers answer strict for it and
// say so, instead of attributing the profile to a pin nobody wrote.
func (c *Config) LoadFailed() bool {
	return c != nil && c.loadFailed
}

// markResolved stamps a Config as produced by a successful Load.
func markResolved(cfg *Config) *Config {
	cfg.resolvedByLoad = true
	return cfg
}

// ResolvedByLoad reports whether this Config came from a successful Load rather
// than being fabricated. Nil-safe: a nil Config was never loaded.
//
// This is the value-level guard behind the profile tail. Every resolver treats a
// Config that is nil, zero-valued, or hand-built on an error path as "the
// profile is unknowable" and answers strict, so a surface that swallows a load
// error cannot silently inherit the loosest default. See config/profile_load.go.
func (c *Config) ResolvedByLoad() bool {
	return c != nil && c.resolvedByLoad
}
