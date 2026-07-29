package scaffold

import (
	"strings"
	"testing"
)

// The scaffold assets are what `centinela init` writes into a NEW project, and
// the binary pins planner-v1 there. An asset that still names the retired pair
// as the plan roles ships stale governance docs to every new project.
//
// This asserts the EMBEDDED bytes, not the file on disk: the embed.FS is what
// actually ships, and workflow-enforcement.md sits on the parity test's
// mirrorParityAllowlist, so byte-parity never covered it.
func TestEmbeddedArchAssets_NamePlannerNotRetiredPlanRoles(t *testing.T) {
	files, err := ListAssetFiles("docs/architecture", ".md")
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no embedded architecture assets found")
	}
	// Phrases that assert the RETIRED roles are the plan step's required roles.
	// A legacy back-compat mention is fine; naming them as the live plan roles
	// is the regression.
	stale := []string{
		"`plan` evidence from `big-thinker` and",
		"- `big-thinker` and `feature-specialist` outputs must include",
	}
	for _, f := range files {
		data, err := ReadAsset(f)
		if err != nil {
			t.Fatalf("read asset %s: %v", f, err)
		}
		body := string(data)
		for _, phrase := range stale {
			if strings.Contains(body, phrase) {
				t.Errorf("embedded asset %s still names the retired plan roles: %q", f, phrase)
			}
		}
	}
}

// The plan-role governance doc must reach new projects naming planner.
func TestEmbeddedWorkflowEnforcement_NamesPlanner(t *testing.T) {
	data, err := ReadAsset("docs/architecture/workflow-enforcement.md")
	if err != nil {
		t.Fatalf("read embedded workflow-enforcement.md: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "`planner`") {
		t.Fatal("embedded workflow-enforcement.md never names the planner role")
	}
	if !strings.Contains(body, "planner-v1") {
		t.Fatal("embedded workflow-enforcement.md does not mention the planner-v1 contract")
	}
}
