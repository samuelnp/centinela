package setup

import (
	"os"
	"testing"
)

// TestAiderAdapterPlanItemsErrors exercises both error returns in
// aiderAdapter.PlanItems: planAgentsFile (AGENTS.md unreadable) and
// planAiderConfig (.aider.conf.yml unreadable). A directory at the target path
// makes os.ReadFile fail with a non-IsNotExist error, hitting the err arm.
func TestAiderAdapterPlanItemsErrors(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	a := aiderAdapter{}

	os.Chdir(t.TempDir())         //nolint:errcheck
	os.MkdirAll(agentsFile, 0755) //nolint:errcheck
	if _, err := a.PlanItems(); err == nil {
		t.Fatal("expected planAgentsFile error when AGENTS.md is a directory")
	}

	os.Chdir(t.TempDir())              //nolint:errcheck
	os.MkdirAll(aiderConfigFile, 0755) //nolint:errcheck
	if _, err := a.PlanItems(); err == nil {
		t.Fatal("expected planAiderConfig error when .aider.conf.yml is a directory")
	}
}

// TestOpenCodeAdapterPlanItemsErrors exercises the planPluginFile and
// planAgentsFile error arms in openCodeAdapter.PlanItems.
func TestOpenCodeAdapterPlanItemsErrors(t *testing.T) {
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	a := openCodeAdapter{}

	os.Chdir(t.TempDir())         //nolint:errcheck
	os.MkdirAll(pluginFile, 0755) //nolint:errcheck
	if _, err := a.PlanItems(); err == nil {
		t.Fatal("expected planPluginFile error when plugin path is a directory")
	}

	os.Chdir(t.TempDir())         //nolint:errcheck
	os.MkdirAll(agentsFile, 0755) //nolint:errcheck
	if _, err := a.PlanItems(); err == nil {
		t.Fatal("expected planAgentsFile error when AGENTS.md is a directory")
	}
}
