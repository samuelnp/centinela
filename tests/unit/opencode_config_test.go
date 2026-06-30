package unit_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestInjectOpenCodeConfig_MergesWithoutLosingKeys(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	seed := `{"command":{"test":{"description":"keep"}},"instructions":["RULES.md"]}`
	os.WriteFile("opencode.json", []byte(seed), 0644) //nolint:errcheck

	changed, err := setup.InjectOpenCodeConfig("opencode.json", nil)
	if err != nil {
		t.Fatalf("InjectOpenCodeConfig error: %v", err)
	}
	if !changed {
		t.Fatal("expected config to be updated")
	}

	data, _ := os.ReadFile("opencode.json") //nolint:errcheck
	var parsed map[string]json.RawMessage
	json.Unmarshal(data, &parsed) //nolint:errcheck

	var schema string
	json.Unmarshal(parsed["$schema"], &schema) //nolint:errcheck
	if schema != "https://opencode.ai/config.json" {
		t.Fatalf("unexpected schema: %q", schema)
	}
	var ins []string
	json.Unmarshal(parsed["instructions"], &ins) //nolint:errcheck
	if len(ins) != 3 || ins[0] != "RULES.md" || ins[1] != "AGENTS.md" || ins[2] != "CLAUDE.md" {
		t.Fatalf("unexpected instructions: %#v", ins)
	}
	if _, ok := parsed["command"]; !ok {
		t.Fatal("existing command key should be preserved")
	}
	var agents map[string]json.RawMessage
	json.Unmarshal(parsed["agent"], &agents) //nolint:errcheck
	for _, name := range []string{"big-thinker", "feature-specialist", "senior-engineer", "qa-senior", "documentation-specialist", "validation-specialist", "ux-ui-specialist"} {
		var cfg map[string]string
		json.Unmarshal(agents[name], &cfg) //nolint:errcheck
		if cfg["mode"] != "subagent" {
			t.Fatalf("expected %s subagent config, got %#v", name, cfg)
		}
	}
}

func TestInjectOpenCodeConfig_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.InjectOpenCodeConfig("opencode.json", nil) //nolint:errcheck
	changed, err := setup.InjectOpenCodeConfig("opencode.json", nil)
	if err != nil {
		t.Fatalf("InjectOpenCodeConfig error: %v", err)
	}
	if changed {
		t.Fatal("expected no changes on already-configured file")
	}
}

func TestInjectOpenCodeConfig_PreservesExistingAgents(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	seed := `{"agent":{"custom":{"description":"keep","mode":"subagent"},"big-thinker":{"description":"mine","mode":"subagent","prompt":"custom"},"build":{"permission":{"task":{"custom":"allow"}}}}}`
	os.WriteFile("opencode.json", []byte(seed), 0644) //nolint:errcheck
	setup.InjectOpenCodeConfig("opencode.json", nil)       //nolint:errcheck

	data, _ := os.ReadFile("opencode.json") //nolint:errcheck
	var parsed map[string]json.RawMessage
	json.Unmarshal(data, &parsed) //nolint:errcheck
	var agents map[string]json.RawMessage
	json.Unmarshal(parsed["agent"], &agents) //nolint:errcheck
	var big map[string]string
	json.Unmarshal(agents["big-thinker"], &big) //nolint:errcheck
	if big["prompt"] != "custom" {
		t.Fatalf("expected existing big-thinker preserved, got %#v", big)
	}
	if _, ok := agents["custom"]; !ok {
		t.Fatal("expected custom agent preserved")
	}
}
