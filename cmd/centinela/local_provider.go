package main

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/setup"
)

// localProviderFrom maps a validated [orchestration.local] block to the setup
// layer's LocalProvider, or nil when no local block is declared. It is a pure
// mapping (no validation or decision logic — those live in internal/config), so
// internal/setup keeps importing nothing internal while cmd/ stays a thin
// orchestrator.
func localProviderFrom(cfg *config.Config) *setup.LocalProvider {
	lc, ok := config.LocalProviderConfig(cfg)
	if !ok {
		return nil
	}
	return &setup.LocalProvider{
		Provider:  lc.Provider,
		Endpoint:  lc.Endpoint,
		Model:     lc.Model,
		APIKeyEnv: lc.APIKeyEnv,
	}
}
