package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// Parity: every runner key the domain advertises via AllowedRunnerKeys() must be
// accepted by the config leaf's model_map validator, and an unknown runner key
// must be rejected. Mirrors the existing tier/role allow-list parity pattern so
// the leaf's local allowedRunnerKeys set cannot drift from the domain.
func TestAllowListParity_AllRunnerKeysAcceptedInModelMap(t *testing.T) {
	for _, runner := range orchestration.AllowedRunnerKeys() {
		toml := "[orchestration.model_map.reasoning]\n" + runner + " = \"some-model\"\n"
		if _, err := loadTempConfig(t, toml); err != nil {
			t.Errorf("runner key %q from AllowedRunnerKeys() rejected by config: %v", runner, err)
		}
	}
}

func TestAllowListParity_AllRunnerKeysAcceptedInOverride(t *testing.T) {
	for _, runner := range orchestration.AllowedRunnerKeys() {
		toml := "[orchestration.models]\nsenior-engineer = { " + runner + " = \"some-model\" }\n"
		if _, err := loadTempConfig(t, toml); err != nil {
			t.Errorf("override runner key %q rejected by config: %v", runner, err)
		}
	}
}

func TestAllowListParity_UnknownRunnerKeyRejected(t *testing.T) {
	_, err := loadTempConfig(t, "[orchestration.model_map.reasoning]\ngemini = \"x\"\n")
	if err == nil || !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("expected unknown runner 'gemini' rejected, got %v", err)
	}
}
