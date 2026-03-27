package unit_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestClaudeStatusLineConfigIsInjected(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	path := ".claude/settings.json"
	changed, err := setup.InjectHooks(path)
	if err != nil || !changed {
		t.Fatalf("InjectHooks = (%v, %v), want (true, nil)", changed, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"statusLine"`)) {
		t.Fatalf("missing statusLine in settings: %s", data)
	}
	if !bytes.Contains(data, []byte(`centinela hook statusline`)) {
		t.Fatalf("missing statusline command in settings: %s", data)
	}
}
