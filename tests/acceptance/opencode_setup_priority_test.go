package acceptance_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/opencode-setup-priority.feature
func TestOpenCodeSetupPriority_GeneratedAssetsFavorSetup(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.InjectOpenCodeConfig("opencode.json") //nolint:errcheck
	setup.EnsureAgentsFile()                    //nolint:errcheck
	setup.EnsureOpenCodePlugin()                //nolint:errcheck

	config, _ := os.ReadFile("opencode.json") //nolint:errcheck
	var parsed map[string]json.RawMessage
	json.Unmarshal(config, &parsed) //nolint:errcheck
	var instructions []string
	json.Unmarshal(parsed["instructions"], &instructions) //nolint:errcheck
	if len(instructions) != 2 || instructions[0] != "AGENTS.md" || instructions[1] != "CLAUDE.md" {
		t.Fatalf("unexpected instructions: %#v", instructions)
	}
	agents, _ := os.ReadFile("AGENTS.md") //nolint:errcheck
	if !strings.Contains(string(agents), "do not reply to greetings first") {
		t.Fatal("expected AGENTS.md to require setup before greeting replies")
	}
	plugin, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	if !strings.Contains(string(plugin), `prependContext(output, joinText(`) {
		t.Fatal("expected plugin to prepend setup directives")
	}
}
