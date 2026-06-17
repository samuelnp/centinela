package config

import (
	"os"
	"testing"
)

func TestAccessors_NilSafe(t *testing.T) {
	if OrchestrationModels(nil) != nil || OrchestrationModelTiers(nil) != nil ||
		OrchestrationModelOverrides(nil) != nil || OrchestrationModelMap(nil) != nil {
		t.Error("accessors must be nil-safe and return nil for a nil config")
	}
}

func TestMixedForms_TierAndTableCoexist(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(dir)                        //nolint:errcheck
	body := "[orchestration.models]\nqa-senior = \"balanced\"\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n"
	os.WriteFile(Filename, []byte(body), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("mixed forms rejected: %v", err)
	}
	if got := OrchestrationModelTiers(cfg)["qa-senior"]; got != "balanced" {
		t.Errorf("qa-senior tier = %q, want balanced", got)
	}
	if got := OrchestrationModelOverrides(cfg)["senior-engineer"]["opencode"]; got != "deepseek/deepseek-coder" {
		t.Errorf("senior-engineer override = %q", got)
	}
	// Back-compat alias returns only the tier form.
	if got := OrchestrationModels(cfg)["qa-senior"]; got != "balanced" {
		t.Errorf("OrchestrationModels alias = %q, want balanced", got)
	}
}

func TestModelMapAccessor_ReturnsTable(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(dir)                        //nolint:errcheck
	body := "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n"
	os.WriteFile(Filename, []byte(body), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := OrchestrationModelMap(cfg)["reasoning"]["opencode"]; got != "moonshotai/kimi-k2" {
		t.Errorf("model_map accessor = %q", got)
	}
}
