package acceptance_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Acceptance: specs/opencode-hook-parity.feature
// Scenario: Plugin invokes prewrite and postwrite around file edits.
func TestOpenCodePlugin_HasPreAndPostWriteHooks(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.EnsureOpenCodePlugin()                             //nolint:errcheck
	data, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	plugin := string(data)

	if !strings.Contains(plugin, `runHook("prewrite"`) {
		t.Fatal("plugin must run prewrite hook")
	}
	if !strings.Contains(plugin, `runHook("postwrite"`) {
		t.Fatal("plugin must run postwrite hook")
	}
}

// Scenario: Plugin invokes setup and context on prompt submit.
func TestOpenCodePlugin_HasSetupAndContextPromptHooks(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.EnsureOpenCodePlugin()                             //nolint:errcheck
	data, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	plugin := string(data)

	if !strings.Contains(plugin, `runHook("setup"`) {
		t.Fatal("plugin must run setup hook")
	}
	if !strings.Contains(plugin, `runHook("context"`) {
		t.Fatal("plugin must run context hook")
	}
}
