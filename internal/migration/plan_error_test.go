package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPlanBuildItemError(t *testing.T) {
	// A managed path that exists as a directory makes buildItem's ReadFile fail
	// with a non-IsNotExist error, which BuildPlan must propagate.
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "CLAUDE.md"), 0755) //nolint:errcheck
	if _, err := BuildPlan(d); err == nil {
		t.Fatal("expected BuildPlan to propagate buildItem read error")
	}
}

func TestBuildItemTemplateReadError(t *testing.T) {
	// A path that is not an embedded scaffold asset makes ReadAsset fail, which
	// buildItem wraps as a "read template" error before touching the filesystem.
	if _, _, err := buildItem(t.TempDir(), "no-such-template.md"); err == nil {
		t.Fatal("expected ReadAsset error for unknown template path")
	}
}

func TestBuildItemLegacyExistingUpdate(t *testing.T) {
	// An existing file with no recognizable header migrates as an update with
	// FromVersion "legacy" (the ParseHeader-fails branch through mergeContent).
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "CLAUDE.md"), []byte("# legacy doc\n"), 0644) //nolint:errcheck
	item, ok, err := buildItem(d, "CLAUDE.md")
	if err != nil || !ok {
		t.Fatalf("expected update item, ok=%v err=%v", ok, err)
	}
	if item.Action != ActionUpdate || item.FromVersion != "legacy" {
		t.Fatalf("expected legacy update, got %#v", item)
	}
	if !strings.Contains(item.content, "centinela:doc-version") {
		t.Fatal("expected migrated content to carry a version header")
	}
}
