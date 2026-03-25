package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestEnsureOpenCodePlugin_IncludesParityHooks(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	changed, err := setup.EnsureOpenCodePlugin()
	if err != nil {
		t.Fatalf("EnsureOpenCodePlugin error: %v", err)
	}
	if !changed {
		t.Fatal("expected plugin to be created")
	}

	data, _ := os.ReadFile(".opencode/plugins/centinela.js") //nolint:errcheck
	s := string(data)
	wants := []string{
		`"tool.execute.before"`,
		`"tool.execute.after"`,
		`"tui.prompt.append"`,
		`normalizeTool(input)`,
		`runHook("prewrite"`,
		`runHook("postwrite"`,
		`runHook("setup"`,
		`runHook("migrate"`,
		`runHook("autostart"`,
		`runHook("orchestration"`,
		`runHook("context"`,
		`const promptPayload = typeof _input === "string" ? _input : JSON.stringify(_input || {})`,
		`args.filePath ||`,
		`args.file_path ||`,
		`args.path ||`,
		`args.filename ||`,
		`nested.filePath ||`,
	}
	for _, w := range wants {
		if !strings.Contains(s, w) {
			t.Fatalf("plugin missing %q", w)
		}
	}
}

func TestEnsureOpenCodePlugin_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	setup.EnsureOpenCodePlugin() //nolint:errcheck
	if changed, _ := setup.EnsureOpenCodePlugin(); changed {
		t.Fatal("expected second call to be no-op")
	}
}
