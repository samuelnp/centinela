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
	if len(instructions) != 1 || instructions[0] != "CLAUDE.md" {
		t.Fatalf("expected CLAUDE.md instruction, got %#v", instructions)
	}
}
