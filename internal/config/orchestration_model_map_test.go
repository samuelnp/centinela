package config

import (
	"os"
	"strings"
	"testing"
)

func loadModelMapTOML(t *testing.T, body string) error {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })       //nolint:errcheck
	os.Chdir(dir)                              //nolint:errcheck
	os.WriteFile(Filename, []byte(body), 0644) //nolint:errcheck
	_, err := Load()
	return err
}

func TestModelMap_ValidEntry(t *testing.T) {
	body := "[orchestration.model_map.reasoning]\nopencode = \"moonshotai/kimi-k2\"\n"
	if err := loadModelMapTOML(t, body); err != nil {
		t.Fatalf("valid model_map rejected: %v", err)
	}
}

func TestModelMap_UnknownRunnerRejected(t *testing.T) {
	body := "[orchestration.model_map.reasoning]\ngemini = \"gemini-pro\"\n"
	err := loadModelMapTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("expected error naming 'gemini', got %v", err)
	}
}

func TestModelMap_UnknownTierRejected(t *testing.T) {
	body := "[orchestration.model_map.turbo]\nopencode = \"x\"\n"
	err := loadModelMapTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "turbo") {
		t.Fatalf("expected error naming 'turbo', got %v", err)
	}
}

func TestModelMap_EmptyModelRejected(t *testing.T) {
	body := "[orchestration.model_map.reasoning]\nopencode = \"\"\n"
	err := loadModelMapTOML(t, body)
	if err == nil || !strings.Contains(err.Error(), "opencode") {
		t.Fatalf("expected error naming empty 'opencode', got %v", err)
	}
}

func TestModelMap_KeyNormalization(t *testing.T) {
	body := "[orchestration.model_map.\" Reasoning \"]\n\" Opencode \" = \"some-model\"\n"
	if err := loadModelMapTOML(t, body); err != nil {
		t.Fatalf("cased/spaced tier+runner keys should normalize: %v", err)
	}
}

func TestModelMap_AbsentAndEmptyValid(t *testing.T) {
	if err := loadModelMapTOML(t, ""); err != nil {
		t.Fatalf("absent model_map: %v", err)
	}
	if err := loadModelMapTOML(t, "[orchestration.model_map]\n"); err != nil {
		t.Fatalf("empty model_map: %v", err)
	}
}

func TestAllowedRunnerKeysList_Format(t *testing.T) {
	got := allowedRunnerKeysList()
	for _, want := range []string{"claude", "opencode", "codex"} {
		if !strings.Contains(got, want) {
			t.Errorf("allowedRunnerKeysList missing %q: %q", want, got)
		}
	}
}
