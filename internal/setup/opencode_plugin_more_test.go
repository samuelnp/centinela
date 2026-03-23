package setup

import (
	"os"
	"testing"
)

func TestEnsureOpenCodePluginMkdirConflict(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile(".opencode", []byte("x"), 0644) //nolint:errcheck
	if _, err := EnsureOpenCodePlugin(); err == nil {
		t.Fatal("expected error when .opencode is a file")
	}
}
