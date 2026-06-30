package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/codex-support.feature

// Scenario: Pre-existing unmanaged .codex/config.toml is not clobbered
func TestAccCodexUnmanagedNotClobbered(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gitInit(t, dir)
	if err := os.MkdirAll(filepath.Join(dir, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "# hand-written, no managed header\nmodel = \"gpt\"\n"
	cfgPath := filepath.Join(dir, ".codex", "config.toml")
	if err := os.WriteFile(cfgPath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := runCent(t, bin, dir, "init", "--agent", "codex")
	if code != 0 {
		t.Fatalf("init exited %d: %s", code, out)
	}
	got, _ := os.ReadFile(cfgPath)
	if string(got) != custom {
		t.Fatalf("unmanaged config was overwritten:\n%s", got)
	}
	if !strings.Contains(out, "manual-review") {
		t.Fatalf("expected manual-review surfaced, got:\n%s", out)
	}
}

// Scenario: relative apply_patch prewrite blocks (regression guard).
func TestAccCodexRelativeApplyPatchBlocks(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gitInit(t, dir)
	src, err := os.ReadFile(filepath.Join(repoRoot(t), "centinela.toml"))
	if err != nil {
		t.Fatalf("read project centinela.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "centinela.toml"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"tool_input":{"command":"*** Begin Patch\n*** Add File: internal/foo.go\n*** End Patch"}}`
	c := exec.Command(bin, "hook", "prewrite")
	c.Dir = dir
	c.Stdin = strings.NewReader(payload)
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run hook prewrite: %v", err)
	}
	if code != 2 {
		t.Fatalf("relative apply_patch code write must exit 2, got %d:\n%s", code, out)
	}
}
