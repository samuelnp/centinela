package config

// PrecommitConfig controls the `centinela precommit` git-hook gate run.
// Enabled governs only the installer/CI advisory surface; the command itself
// runs whenever invoked (the installed hook is the opt-in). SkipBuild (default
// true) drops the heavy cross-compile build gate so the hook stays fast.
//
// SkipBuild is decoded via the *bool RawSkipBuild so an omitted skip_build
// (nil → default true) is distinguishable from an explicit skip_build = false.
type PrecommitConfig struct {
	Enabled      bool  `toml:"enabled"`
	SkipBuild    bool  `toml:"-"`
	RawSkipBuild *bool `toml:"skip_build"`
}

// NormalizePrecommit resolves SkipBuild from the raw decoded pointer: nil
// (section/field omitted) defaults to true; an explicit value is honored.
func NormalizePrecommit(c PrecommitConfig) PrecommitConfig {
	if c.RawSkipBuild == nil {
		c.SkipBuild = true
	} else {
		c.SkipBuild = *c.RawSkipBuild
	}
	return c
}

// validatePrecommit is reserved for future knobs; currently a no-op.
func validatePrecommit(_ PrecommitConfig) error {
	return nil
}
