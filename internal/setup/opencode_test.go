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

func TestEnsureOpenCodePluginErrorWhenPathConflicts(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                            //nolint:errcheck
	os.Chdir(d)                                  //nolint:errcheck
	os.WriteFile(".opencode", []byte("x"), 0644) //nolint:errcheck
	if _, err := EnsureOpenCodePlugin(); err == nil {
		t.Fatal("expected plugin creation error for conflicting .opencode file")
	}
}

func TestEnsureAgentsFileExistingNoChange(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                            //nolint:errcheck
	os.Chdir(d)                                  //nolint:errcheck
	os.WriteFile("AGENTS.md", []byte("x"), 0644) //nolint:errcheck
	if changed, err := EnsureAgentsFile(); err != nil || changed {
		t.Fatalf("EnsureAgentsFile existing = (%v, %v)", changed, err)
	}
}

func TestEnsureOpenCodePluginExistingNoChange(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                 //nolint:errcheck
	os.Chdir(d)                                                       //nolint:errcheck
	os.MkdirAll(".opencode/plugins", 0755)                            //nolint:errcheck
	os.WriteFile(".opencode/plugins/centinela.js", []byte("x"), 0644) //nolint:errcheck
	if changed, err := EnsureOpenCodePlugin(); err != nil || changed {
		t.Fatalf("EnsureOpenCodePlugin existing = (%v, %v)", changed, err)
	}
}
