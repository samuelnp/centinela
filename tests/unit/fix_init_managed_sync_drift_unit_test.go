package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

// Unit: a clean repo's OpenCode sync plan creates the managed plugin + AGENTS.md
// (the assets init must write in their migrated, header'd form).
func TestOpencodeSyncPlanCreatesManagedAssets(t *testing.T) {
	t.Chdir(t.TempDir())
	plan, err := setup.BuildSyncPlan("opencode")
	if err != nil {
		t.Fatal(err)
	}
	action := map[string]setup.SyncAction{}
	for _, it := range plan.Items {
		action[it.Path] = it.Action
	}
	for _, p := range []string{"AGENTS.md", ".opencode/plugins/centinela.js"} {
		if action[p] != setup.SyncCreate {
			t.Errorf("%s: action=%q want create", p, action[p])
		}
	}
}
