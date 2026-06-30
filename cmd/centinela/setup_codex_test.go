package main

import (
	"os"
	"strings"
	"testing"
)

func TestSetupCodex_WritesManagedFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := setupCodex(); err != nil {
		t.Fatalf("setupCodex: %v", err)
	}
	cfg, err := os.ReadFile(".codex/config.toml")
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.HasPrefix(string(cfg), "# centinela:managed-version=") {
		t.Fatalf("config missing managed header:\n%s", cfg)
	}
	if !strings.Contains(string(cfg), "centinela hook prewrite") {
		t.Fatalf("config missing prewrite hook:\n%s", cfg)
	}
	if _, err := os.Stat("AGENTS.md"); err != nil {
		t.Fatalf("AGENTS.md not written: %v", err)
	}
}

func TestSetupCodex_UnmanagedNotClobbered(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".codex", 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "# my hand-written codex config\nmodel = \"gpt\"\n"
	if err := os.WriteFile(".codex/config.toml", []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if err := setupCodex(); err != nil {
			t.Fatalf("setupCodex: %v", err)
		}
	})
	got, _ := os.ReadFile(".codex/config.toml")
	if string(got) != custom {
		t.Fatalf("unmanaged config was clobbered:\n%s", got)
	}
	if !strings.Contains(out, "manual-review") {
		t.Fatalf("expected manual-review warning, got:\n%s", out)
	}
}
