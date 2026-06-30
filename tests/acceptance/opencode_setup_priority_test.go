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

	setup.InjectOpenCodeConfig("opencode.json", nil) //nolint:errcheck
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
	if !strings.Contains(string(agents), "On a greeting-only first prompt") {
		t.Fatal("expected AGENTS.md to cover greeting-only first prompts")
	}
	if !strings.Contains(string(agents), "do not suggest centinela start <feature>") {
		t.Fatal("expected AGENTS.md to block feature workflow suggestions during setup")
	}
	if !strings.Contains(string(agents), "do not ask what to work on") {
		t.Fatal("expected AGENTS.md to block feature discovery during setup")
	}
	if !strings.Contains(string(agents), "define the roadmap before asking for feature work") {
		t.Fatal("expected AGENTS.md to require roadmap bootstrap before feature work")
	}
	plugin, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	if !strings.Contains(string(plugin), `prependContext(output, joinText(`) {
		t.Fatal("expected plugin to prepend setup directives")
	}
}
