package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// localProviderFrom: nil cfg and an unset local block both map to nil; a set local
// block maps to a *setup.LocalProvider carrying the normalized/trimmed fields.
func TestLocalProviderFrom(t *testing.T) {
	if localProviderFrom(nil) != nil {
		t.Fatal("nil cfg must map to nil")
	}
	if localProviderFrom(&config.Config{}) != nil {
		t.Fatal("no local block must map to nil")
	}
	cfg := &config.Config{}
	cfg.Orchestration.Local = config.LocalConfig{Provider: "  Ollama  ", Endpoint: "  http://x/v1  ", Model: "  m  ", APIKeyEnv: "  K  "}
	lp := localProviderFrom(cfg)
	if lp == nil {
		t.Fatal("a set local block must map to a provider")
	}
	if lp.Provider != "ollama" || lp.Endpoint != "http://x/v1" || lp.Model != "m" || lp.APIKeyEnv != "K" {
		t.Fatalf("mapping not normalized: %+v", lp)
	}
}
