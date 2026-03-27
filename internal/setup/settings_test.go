package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInjectHooksCreatesAndPreserves(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	path := filepath.Join(".claude", "settings.json")
	if changed, err := InjectHooks(path); err != nil || !changed {
		t.Fatalf("InjectHooks #1 = (%v, %v)", changed, err)
	}
	if changed, err := InjectHooks(path); err != nil || changed {
		t.Fatalf("InjectHooks #2 = (%v, %v)", changed, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing settings file: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"statusLine"`)) {
		t.Fatalf("expected statusLine config, got: %s", string(data))
	}
}
