package setup

import (
	"os"
	"strings"
	"testing"
)

func TestPlanAiderConfig_Create(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	item, err := planAiderConfig()
	if err != nil {
		t.Fatalf("planAiderConfig: %v", err)
	}
	if item == nil || item.Action != SyncCreate {
		t.Fatalf("expected create item, got %v", item)
	}
	if item.Kind != SyncAiderConfig {
		t.Fatalf("expected SyncAiderConfig kind, got %s", item.Kind)
	}
}

func TestPlanAiderConfig_Idempotent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	target := aiderConfigHeader + "\n" + aiderConfigBody
	os.WriteFile(aiderConfigFile, []byte(target), 0644) //nolint:errcheck

	item, err := planAiderConfig()
	if err != nil {
		t.Fatalf("planAiderConfig: %v", err)
	}
	if item != nil {
		t.Fatalf("expected nil (no-op), got item action=%s", item.Action)
	}
}

func TestPlanAiderConfig_UpdateManaged(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	// Managed file with stale content triggers update.
	os.WriteFile(aiderConfigFile, []byte("# centinela:managed-version=0\nread: old\n"), 0644) //nolint:errcheck

	item, err := planAiderConfig()
	if err != nil {
		t.Fatalf("planAiderConfig: %v", err)
	}
	if item == nil || item.Action != SyncUpdate {
		t.Fatalf("expected update item, got %v", item)
	}
}

func TestPlanAiderConfig_ManualReviewForUnmanaged(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile(aiderConfigFile, []byte("# user custom config\nread: myfile\n"), 0644) //nolint:errcheck

	item, err := planAiderConfig()
	if err != nil {
		t.Fatalf("planAiderConfig: %v", err)
	}
	if item == nil || item.Action != SyncManualReview {
		t.Fatalf("expected manual-review, got %v", item)
	}
}

func TestWriteManagedAiderConfig_Content(t *testing.T) {
	d := t.TempDir()
	path := d + "/" + aiderConfigFile
	if err := writeManagedAiderConfig(path); err != nil {
		t.Fatalf("writeManagedAiderConfig: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "read: AGENTS.md") {
		t.Fatalf("missing read: AGENTS.md in:\n%s", content)
	}
	if !strings.Contains(content, "centinela:managed-version=") {
		t.Fatalf("missing managed header in:\n%s", content)
	}
}
