package main

import "github.com/samuelnp/centinela/internal/config"

// precommitCfg returns the config used for the precommit gate run. When
// cfg.Precommit.SkipBuild is set (the default), it returns a shallow value copy
// with the heavy cross-compile build gate disabled — leaving the caller's cfg
// unmutated — so the hook stays fast. Otherwise the original cfg is returned.
func precommitCfg(cfg *config.Config) *config.Config {
	if !cfg.Precommit.SkipBuild {
		return cfg
	}
	c := *cfg
	c.Gates.Build.Enabled = false
	return &c
}
