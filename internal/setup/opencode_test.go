package setup

import (
	"os"
	"testing"
)

func TestOpenCodeSetupFiles(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if changed, err := InjectOpenCodeConfig("opencode.json"); err != nil || !changed {
		t.Fatalf("InjectOpenCodeConfig: %v %v", changed, err)
	}
	if changed, err := EnsureOpenCodePlugin(); err != nil || !changed {
		t.Fatalf("EnsureOpenCodePlugin: %v %v", changed, err)
	}
	if changed, err := EnsureAgentsFile(); err != nil || !changed {
		t.Fatalf("EnsureAgentsFile: %v %v", changed, err)
	}
	if changed, _ := InjectOpenCodeConfig("opencode.json"); changed {
		t.Fatal("expected idempotent OpenCode config")
	}
}
