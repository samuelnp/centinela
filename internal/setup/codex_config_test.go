package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	os.Chdir(d)                       //nolint:errcheck
}

func TestPlanCodexConfig_Create(t *testing.T) {
	chdirTemp(t)
	item, err := planCodexConfig()
	if err != nil {
		t.Fatalf("planCodexConfig: %v", err)
	}
	if item == nil || item.Action != SyncCreate {
		t.Fatalf("expected create item, got %v", item)
	}
	if item.Kind != SyncKindPrewriteHook {
		t.Fatalf("expected SyncKindPrewriteHook, got %s", item.Kind)
	}
}

func TestPlanCodexConfig_UpdateManaged(t *testing.T) {
	chdirTemp(t)
	writeCodex(t, "# centinela:managed-version=0 template=.codex/config.toml\nold = 1\n")
	item, err := planCodexConfig()
	if err != nil {
		t.Fatalf("planCodexConfig: %v", err)
	}
	if item == nil || item.Action != SyncUpdate {
		t.Fatalf("expected update item, got %v", item)
	}
}

func TestPlanCodexConfig_ManualReviewForUnmanaged(t *testing.T) {
	chdirTemp(t)
	writeCodex(t, "# hand-written codex config\nmodel = \"gpt\"\n")
	item, err := planCodexConfig()
	if err != nil {
		t.Fatalf("planCodexConfig: %v", err)
	}
	if item == nil || item.Action != SyncManualReview {
		t.Fatalf("expected manual-review (never clobbered), got %v", item)
	}
}

func TestWriteManagedCodexConfig_HeaderAndBody(t *testing.T) {
	d := t.TempDir()
	path := filepath.Join(d, codexConfigFile)
	if err := writeManagedCodexConfig(path); err != nil {
		t.Fatalf("writeManagedCodexConfig: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "# centinela:managed-version=") {
		t.Fatalf("missing managed-version header:\n%s", content)
	}
	if !strings.Contains(content, "apply_patch") {
		t.Fatalf("missing apply_patch matcher:\n%s", content)
	}
}

func writeCodex(t *testing.T, body string) {
	t.Helper()
	if err := os.MkdirAll(".codex", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codexConfigFile, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
