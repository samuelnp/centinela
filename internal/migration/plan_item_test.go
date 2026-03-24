package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/scaffold"
)

func TestBuildItemNoopWhenVersionMatches(t *testing.T) {
	d := t.TempDir()
	tpl, err := scaffold.ReadAsset("CLAUDE.md")
	if err != nil {
		t.Fatal(err)
	}
	content := WithHeader(string(tpl), "CLAUDE.md", CurrentDocVersion)
	os.WriteFile(filepath.Join(d, "CLAUDE.md"), []byte(content), 0644) //nolint:errcheck
	_, ok, err := buildItem(d, "CLAUDE.md")
	if err != nil || ok {
		t.Fatalf("expected no-op item, ok=%v err=%v", ok, err)
	}
}

func TestBuildItemReadError(t *testing.T) {
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "CLAUDE.md"), 0755) //nolint:errcheck
	_, _, err := buildItem(d, "CLAUDE.md")
	if err == nil {
		t.Fatal("expected read error for directory path")
	}
}
