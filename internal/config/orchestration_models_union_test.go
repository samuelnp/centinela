package config

import (
	"os"
	"strings"
	"testing"
)

func loadUnionTOML(t *testing.T, body string) (*Config, error) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })       //nolint:errcheck
	os.Chdir(dir)                              //nolint:errcheck
	os.WriteFile(Filename, []byte(body), 0644) //nolint:errcheck
	return Load()
}

func TestRoleModelValue_UnmarshalString(t *testing.T) {
	var v RoleModelValue
	if err := v.UnmarshalTOML("balanced"); err != nil || v.Tier != "balanced" {
		t.Fatalf("string form: tier=%q err=%v", v.Tier, err)
	}
}

func TestRoleModelValue_UnmarshalTable(t *testing.T) {
	var v RoleModelValue
	if err := v.UnmarshalTOML(map[string]any{"opencode": "deepseek/deepseek-coder"}); err != nil {
		t.Fatalf("table form err: %v", err)
	}
	if v.Overrides["opencode"] != "deepseek/deepseek-coder" {
		t.Errorf("override not captured: %v", v.Overrides)
	}
}

func TestRoleModelValue_UnmarshalNonStringValueErrors(t *testing.T) {
	var v RoleModelValue
	if err := v.UnmarshalTOML(map[string]any{"opencode": 42}); err == nil {
		t.Fatal("expected error for non-string model value")
	}
}

func TestRoleModelValue_UnmarshalWrongTypeErrors(t *testing.T) {
	var v RoleModelValue
	if err := v.UnmarshalTOML(42); err == nil {
		t.Fatal("expected error for non-string/non-table value")
	}
}

func TestUnion_RoleOverrideTable(t *testing.T) {
	body := "[orchestration.models]\nsenior-engineer = { opencode = \"deepseek/deepseek-coder\" }\n"
	cfg, err := loadUnionTOML(t, body)
	if err != nil {
		t.Fatalf("override table rejected: %v", err)
	}
	if got := OrchestrationModelOverrides(cfg)["senior-engineer"]["opencode"]; got != "deepseek/deepseek-coder" {
		t.Errorf("override accessor wrong: %q", got)
	}
}

func TestUnion_UnknownRoleTableRejected(t *testing.T) {
	body := "[orchestration.models]\nbackend-wizard = { opencode = \"x\" }\n"
	_, err := loadUnionTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "backend-wizard") {
		t.Fatalf("expected error naming 'backend-wizard', got %v", err)
	}
}

func TestUnion_OverrideUnknownRunnerRejected(t *testing.T) {
	body := "[orchestration.models]\nsenior-engineer = { gemini = \"x\" }\n"
	_, err := loadUnionTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("expected error naming 'gemini', got %v", err)
	}
}

func TestUnion_OverrideEmptyModelRejected(t *testing.T) {
	body := "[orchestration.models]\nsenior-engineer = { opencode = \"\" }\n"
	_, err := loadUnionTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "opencode") {
		t.Fatalf("expected error naming empty 'opencode', got %v", err)
	}
}
