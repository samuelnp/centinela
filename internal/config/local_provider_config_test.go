package config

import "testing"

// LocalProviderConfig: nil-safe; unset → zero+false; normalizes provider
// (trim+lower) and trims the opaque endpoint/model/api_key_env fields; ok is true
// only when the provider is non-empty after normalization.
func TestLocalProviderConfig(t *testing.T) {
	if lc, ok := LocalProviderConfig(nil); ok || lc.Provider != "" {
		t.Fatalf("nil cfg: got (%+v,%v)", lc, ok)
	}
	if lc, ok := LocalProviderConfig(&Config{}); ok || lc.Provider != "" {
		t.Fatalf("empty cfg: got (%+v,%v)", lc, ok)
	}
	lc, ok := LocalProviderConfig(cfgLocal("  Ollama  ", "  http://x/v1  ", "  m  ", "  K  "))
	if !ok {
		t.Fatal("expected ok for a set provider")
	}
	if lc.Provider != "ollama" {
		t.Fatalf("provider not normalized: %q", lc.Provider)
	}
	if lc.Endpoint != "http://x/v1" || lc.Model != "m" || lc.APIKeyEnv != "K" {
		t.Fatalf("opaque fields not trimmed: %+v", lc)
	}
}

// Back-compat: with no [orchestration.local] block, every new local tier is inert
// — driver resolution, the local capability default, and provider gating are all
// unchanged from the pre-feature zero-config path.
func TestLocalTiersInertWithoutBlock(t *testing.T) {
	cfg := &Config{}
	if got := DriverModelFrom("", cfg); got != "" {
		t.Fatalf("zero-config driver model = %q, want empty", got)
	}
	if _, ok := LocalProviderConfig(cfg); ok {
		t.Fatal("no local block must report unset")
	}
	if _, ok := LocalDefaultClass("any-model", cfg); ok {
		t.Fatal("local default must not engage without a local block")
	}
	if p, ok := DefaultProfileForModel("any-model", cfg); ok || p != "" {
		t.Fatalf("unmapped model without local = (%q,%v), want (\"\",false)", p, ok)
	}
}
