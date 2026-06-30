package config

import "strings"

// LocalDefaultClass is the strictly-lowest capability tier: it classifies the
// declared [orchestration.local].model as `limited` ONLY when that model is the
// configured local model AND it has no explicit/builtin class via
// CapabilityClassFor. It returns ("", false) when the id is empty, cfg is nil,
// the id is not the local model, or the id already resolves to a class — so an
// explicit [orchestration.capabilities] mapping (and the builtin map) always win
// and the zero-config path (no local block) is never engaged.
func LocalDefaultClass(modelID string, cfg *Config) (string, bool) {
	id := strings.TrimSpace(modelID)
	if id == "" || cfg == nil {
		return "", false
	}
	local, ok := LocalProviderConfig(cfg)
	if !ok || local.Model != id {
		return "", false
	}
	if _, classed := CapabilityClassFor(id, cfg); classed {
		return "", false
	}
	return CapabilityLimited, true
}
