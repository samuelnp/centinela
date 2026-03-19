package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestOpenCodePlugin_GeneratesPromptAndPostwriteHooks(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	if _, err := setup.EnsureOpenCodePlugin(); err != nil {
		t.Fatalf("EnsureOpenCodePlugin error: %v", err)
	}

	data, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	plugin := string(data)

	if !strings.Contains(plugin, `"tool.execute.after"`) {
		t.Fatal("expected tool.execute.after hook")
	}
	if !strings.Contains(plugin, `"tui.prompt.append"`) {
		t.Fatal("expected tui.prompt.append hook")
	}
	if !strings.Contains(plugin, `appendContext(output, runHook("setup"`) {
		t.Fatal("expected setup hook output appended to prompt context")
	}
	if !strings.Contains(plugin, `appendContext(output, runHook("context"`) {
		t.Fatal("expected context hook output appended to prompt context")
	}
}
