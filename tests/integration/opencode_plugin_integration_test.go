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
	if !strings.Contains(plugin, `prependContext(output, joinText(`) {
		t.Fatal("expected setup or migrate output prepended to prompt context")
	}
	if strings.Index(plugin, `runHook("setup"`) > strings.Index(plugin, `runHook("autostart"`) {
		t.Fatal("expected setup hook before autostart hook")
	}
	if strings.Index(plugin, `runHook("migrate"`) > strings.Index(plugin, `runHook("autostart"`) {
		t.Fatal("expected migrate hook before autostart hook")
	}
	if !strings.Contains(plugin, `appendContext(output, runHook("autostart"`) {
		t.Fatal("expected autostart hook output appended to prompt context")
	}
	if !strings.Contains(plugin, `appendContext(output, runHook("orchestration"`) {
		t.Fatal("expected orchestration hook output appended to prompt context")
	}
	if !strings.Contains(plugin, `appendContext(output, runHook("context"`) {
		t.Fatal("expected context hook output appended to prompt context")
	}
	if !strings.Contains(plugin, `input.tool || input.toolName || input.name`) {
		t.Fatal("expected normalized tool fallback keys")
	}
	if !strings.Contains(plugin, `output.prompt = front ? text + "\n\n" + output.prompt`) {
		t.Fatal("expected prepended prompt text handling")
	}
}
