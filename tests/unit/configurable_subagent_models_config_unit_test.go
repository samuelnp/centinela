package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

// loadTempConfig writes centinela.toml content to a temp dir and calls config.Load.
func loadTempConfig(t *testing.T, tomlContent string) (*config.Config, error) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })                      //nolint:errcheck
	os.Chdir(d)                                               //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(tomlContent), 0644) //nolint:errcheck
	return config.Load()
}

func TestValidateOrchestrationModels_AbsentTable(t *testing.T) {
	_, err := loadTempConfig(t, "")
	if err != nil {
		t.Fatalf("absent table: expected nil error, got %v", err)
	}
}

func TestValidateOrchestrationModels_EmptyTable(t *testing.T) {
	_, err := loadTempConfig(t, "[orchestration.models]\n")
	if err != nil {
		t.Fatalf("empty table: expected nil error, got %v", err)
	}
}

func TestValidateOrchestrationModels_ValidMapping(t *testing.T) {
	toml := "[orchestration.models]\nbig-thinker = \"reasoning\"\nqa-senior = \"fast\"\n"
	_, err := loadTempConfig(t, toml)
	if err != nil {
		t.Fatalf("valid mapping: expected nil error, got %v", err)
	}
}

func TestValidateOrchestrationModels_UnknownTierRejected(t *testing.T) {
	toml := "[orchestration.models]\nqa-senior = \"genius\"\n"
	_, err := loadTempConfig(t, toml)
	if err == nil {
		t.Fatal("unknown tier: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "qa-senior") {
		t.Errorf("error should name the key 'qa-senior', got: %v", err)
	}
	for _, tier := range []string{"reasoning", "balanced", "fast"} {
		if !strings.Contains(err.Error(), tier) {
			t.Errorf("error should list allowed tier %q, got: %v", tier, err)
		}
	}
}

func TestValidateOrchestrationModels_UnknownRoleRejected(t *testing.T) {
	toml := "[orchestration.models]\nbackend-wizard = \"fast\"\n"
	_, err := loadTempConfig(t, toml)
	if err == nil {
		t.Fatal("unknown role: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "backend-wizard") {
		t.Errorf("error should name the key 'backend-wizard', got: %v", err)
	}
}

func TestValidateOrchestrationModels_CasingNormalized(t *testing.T) {
	toml := "[orchestration.models]\nfeature-specialist = \"Reasoning\"\n"
	cfg, err := loadTempConfig(t, toml)
	if err != nil {
		t.Fatalf("casing normalized: expected nil error, got %v", err)
	}
	if v := config.OrchestrationModels(cfg)["feature-specialist"]; v != "Reasoning" {
		t.Errorf("raw stored value should be preserved as-is; got %q", v)
	}
}

func TestAllowListParity_AllOrchestrationTiersAccepted(t *testing.T) {
	for _, tier := range orchestration.AllowedTiers() {
		toml := "[orchestration.models]\nbig-thinker = \"" + string(tier) + "\"\n"
		if _, err := loadTempConfig(t, toml); err != nil {
			t.Errorf("tier %q from AllowedTiers() rejected by config: %v", tier, err)
		}
	}
}

func TestAllowListParity_AllOrchestrationRoleSlugsAccepted(t *testing.T) {
	for _, slug := range orchestration.AllowedRoleSlugs() {
		toml := "[orchestration.models]\n" + slug + " = \"fast\"\n"
		if _, err := loadTempConfig(t, toml); err != nil {
			t.Errorf("role slug %q from AllowedRoleSlugs() rejected by config: %v", slug, err)
		}
	}
}
