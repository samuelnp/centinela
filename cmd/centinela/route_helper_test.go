package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

const dynamicToml = "[orchestration]\nrouting_mode = \"dynamic\"\n"

// routeRepo chdirs into a temp project holding a saved workflow at step, with
// the given centinela.toml, and resets the --reason flag between tests.
func routeRepo(t *testing.T, step, toml string) *workflow.Workflow {
	t.Helper()
	dir := t.TempDir()
	origin, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origin) })             //nolint:errcheck
	t.Cleanup(func() { routeSetReason = "" })          //nolint:errcheck
	os.Chdir(dir)                                      //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)            //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = step
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	return wf
}

// routeOf reloads the feature's workflow and returns the role's recorded route.
func routeOf(t *testing.T, role string) (workflow.ModelRoute, bool) {
	t.Helper()
	wf, err := workflow.Load("f")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	return wf.ModelRouteFor(role)
}
