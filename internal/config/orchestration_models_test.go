package config

import (
	"os"
	"strings"
	"testing"
)

func TestValidateOrchestrationModels_AbsentEmpty(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	// absent table
	if _, err := Load(); err != nil {
		t.Fatalf("absent table: unexpected error: %v", err)
	}
	// empty table
	os.WriteFile(Filename, []byte("[orchestration.models]\n"), 0644) //nolint:errcheck
	if _, err := Load(); err != nil {
		t.Fatalf("empty table: unexpected error: %v", err)
	}
}

func TestValidateOrchestrationModels_ValidMapping(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.models]\nbig-thinker = \"reasoning\"\nqa-senior = \"fast\"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	if _, err := Load(); err != nil {
		t.Fatalf("valid mapping: unexpected error: %v", err)
	}
}

func TestValidateOrchestrationModels_UnknownTier(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.models]\nqa-senior = \"genius\"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unknown tier, got nil")
	}
	if !strings.Contains(err.Error(), "qa-senior") {
		t.Errorf("error should name key 'qa-senior': %v", err)
	}
	if !strings.Contains(err.Error(), "reasoning") {
		t.Errorf("error should list allowed tiers: %v", err)
	}
}

func TestValidateOrchestrationModels_UnknownRole(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.models]\nbackend-wizard = \"fast\"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unknown role, got nil")
	}
	if !strings.Contains(err.Error(), "backend-wizard") {
		t.Errorf("error should name key 'backend-wizard': %v", err)
	}
}

func TestValidateOrchestrationModels_CasingNormalized(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.models]\nfeature-specialist = \"Reasoning\"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	if _, err := Load(); err != nil {
		t.Fatalf("cased tier should be normalized and accepted: %v", err)
	}
}

func TestValidateOrchestrationModels_InvalidAfterNormalize(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	toml := "[orchestration.models]\nsenior-engineer = \" Genius \"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	if _, err := Load(); err == nil {
		t.Fatal("expected error for ' Genius ' after normalization, got nil")
	}
}
