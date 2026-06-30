package integration_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestOpenCodeSetup_WritesAllArtifacts(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	if changed, err := setup.InjectOpenCodeConfig("opencode.json", nil); err != nil || !changed {
		t.Fatalf("InjectOpenCodeConfig = (%v, %v), want (true, nil)", changed, err)
	}
	if changed, err := setup.EnsureOpenCodePlugin(); err != nil || !changed {
		t.Fatalf("EnsureOpenCodePlugin = (%v, %v), want (true, nil)", changed, err)
	}
	if changed, err := setup.EnsureAgentsFile(); err != nil || !changed {
		t.Fatalf("EnsureAgentsFile = (%v, %v), want (true, nil)", changed, err)
	}

	requireFile(t, "opencode.json")
	requireFile(t, ".opencode/plugins/centinela.js")
	requireFile(t, "AGENTS.md")

	if changed, _ := setup.EnsureOpenCodePlugin(); changed {
		t.Fatal("plugin write should be idempotent")
	}
	if changed, _ := setup.EnsureAgentsFile(); changed {
		t.Fatal("AGENTS.md write should be idempotent")
	}
}

func requireFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing file %s: %v", path, err)
	}
}
