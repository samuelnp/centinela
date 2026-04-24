package acceptance_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Scenario: Existing OpenCode config is preserved.
func TestOpenCodeConfig_ExistingKeysRemain(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	os.WriteFile("opencode.json", []byte(`{"command":{"review":{"description":"keep"}}}`), 0644) //nolint:errcheck
	if _, err := setup.InjectOpenCodeConfig("opencode.json"); err != nil {
		t.Fatalf("InjectOpenCodeConfig error: %v", err)
	}

	data, _ := os.ReadFile("opencode.json") //nolint:errcheck
	var parsed map[string]json.RawMessage
	json.Unmarshal(data, &parsed) //nolint:errcheck

	if _, ok := parsed["command"]; !ok {
		t.Fatal("existing command config should remain")
	}
	var instructions []string
	json.Unmarshal(parsed["instructions"], &instructions) //nolint:errcheck
	if len(instructions) != 2 || instructions[0] != "AGENTS.md" || instructions[1] != "CLAUDE.md" {
		t.Fatalf("expected AGENTS.md and CLAUDE.md instructions, got %#v", instructions)
	}
}
