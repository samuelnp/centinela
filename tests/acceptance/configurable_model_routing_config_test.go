package acceptance_test

// Acceptance: specs/configurable-model-routing.feature (AC4, AC5, normalization).

import (
	"strings"
	"testing"
)

// expectValidateError runs `centinela validate` and asserts the config error
// names the offending key.
func expectValidateError(t *testing.T, toml, key string) {
	t.Helper()
	d, bin := configOnlyDir(t, toml)
	out, _ := runBin(t, bin, d, "validate")
	if !strings.Contains(out, key) {
		t.Errorf("expected validate to name offending key %q; got:\n%s", key, out)
	}
}

// AC5: unknown runner key in model_map fails loudly naming the key.
func TestRoutingConfig_UnknownRunnerInModelMap(t *testing.T) {
	expectValidateError(t, "[orchestration.model_map.reasoning]\ngemini = \"gemini-pro\"\n", "gemini")
}

// AC5: unknown role key in orchestration.models (table form) fails loudly.
func TestRoutingConfig_UnknownRoleInModels(t *testing.T) {
	expectValidateError(t, "[orchestration.models]\nbackend-wizard = { opencode = \"some-model\" }\n", "backend-wizard")
}

// AC5: unknown tier key in model_map fails loudly.
func TestRoutingConfig_UnknownTierInModelMap(t *testing.T) {
	expectValidateError(t, "[orchestration.model_map.turbo]\nopencode = \"x\"\n", "turbo")
}

// AC5: empty model string in model_map fails loudly naming the runner key.
func TestRoutingConfig_EmptyModelInModelMap(t *testing.T) {
	expectValidateError(t, "[orchestration.model_map.reasoning]\nopencode = \"\"\n", "opencode")
}

// AC4: a plain tier string in orchestration.models still loads (back-compat).
func TestRoutingConfig_PlainTierStringBackCompat(t *testing.T) {
	d, bin := configOnlyDir(t, "[orchestration.models]\nqa-senior = \"balanced\"\n")
	out, err := runBin(t, bin, d, "validate")
	if strings.Contains(out, "unknown") || strings.Contains(out, "invalid tier") {
		t.Errorf("AC4: plain tier string should load cleanly; got:\n%s (err=%v)", out, err)
	}
}

// Edge: mixed forms (tier string + override table) load together.
func TestRoutingConfig_MixedFormsLoad(t *testing.T) {
	toml := "[orchestration.models]\nqa-senior = \"balanced\"\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n"
	d, bin := configOnlyDir(t, toml)
	out, _ := runBin(t, bin, d, "validate")
	if strings.Contains(out, "unknown") || strings.Contains(out, "must not be empty") {
		t.Errorf("edge: mixed forms should load; got:\n%s", out)
	}
}

// Edge: cased/spaced tier and runner keys normalize and load.
func TestRoutingConfig_KeyNormalization(t *testing.T) {
	toml := "[orchestration.model_map.\" Reasoning \"]\n\" Opencode \" = \"some-model\"\n"
	d, bin := configOnlyDir(t, toml)
	out, _ := runBin(t, bin, d, "validate")
	if strings.Contains(out, "unknown tier") || strings.Contains(out, "unknown runner") {
		t.Errorf("edge: cased keys should normalize; got:\n%s", out)
	}
}
