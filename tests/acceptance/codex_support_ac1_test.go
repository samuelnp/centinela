package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/codex-support.feature

// Scenario: centinela init --agent codex writes managed .codex/config.toml
func TestAccCodexInitWritesManagedConfig(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gitInit(t, dir)
	if out, code := runCent(t, bin, dir, "init", "--agent", "codex"); code != 0 {
		t.Fatalf("init exited %d: %s", code, out)
	}
	cfg, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	s := string(cfg)
	for _, want := range []string{
		"# centinela:managed-version=", "apply_patch",
		"centinela hook prewrite", "centinela hook postwrite", "UserPromptSubmit",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, s)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Fatalf(".claude/settings.json should NOT exist for codex init")
	}
}

// Scenario: init then migrate setup reports no pending drift
func TestAccCodexInitMigrateNoDrift(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gitInit(t, dir)
	if out, code := runCent(t, bin, dir, "init", "--agent", "codex"); code != 0 {
		t.Fatalf("init exited %d: %s", code, out)
	}
	out, code := runCent(t, bin, dir, "migrate", "setup", "--agent", "codex")
	if code != 0 {
		t.Fatalf("migrate exited %d: %s", code, out)
	}
	if strings.Contains(out, "create:") || strings.Contains(out, "update:") {
		t.Fatalf("fresh codex init must leave no drift, got:\n%s", out)
	}
}
