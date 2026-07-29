package evidence

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func retiredTempRepo(t *testing.T) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

func saveContract(t *testing.T, feature, contract string) {
	t.Helper()
	wf := workflow.New(feature)
	wf.PlanContract = contract
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
}

// Legacy in-flight workflow: the pair it must still write stays authorable.
func TestEnsureRoleAllowed_LegacyWorkflowAcceptsRetiredRoles(t *testing.T) {
	retiredTempRepo(t)
	saveContract(t, "old", "")
	for _, r := range []Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial} {
		if err := EnsureRoleAllowed("old", r); err != nil {
			t.Errorf("legacy workflow must still author %q: %v", r, err)
		}
	}
}

func TestEnsureRoleAllowed_PinnedWorkflowRefusesRetiredRoles(t *testing.T) {
	retiredTempRepo(t)
	saveContract(t, "fresh", workflow.PlanContractUnified)
	for _, r := range []Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial} {
		err := EnsureRoleAllowed("fresh", r)
		if err == nil {
			t.Fatalf("pinned workflow must refuse retired role %q", r)
		}
		if !strings.Contains(err.Error(), "retired") || !strings.Contains(err.Error(), "planner") {
			t.Fatalf("error must say retired and point at planner, got %v", err)
		}
	}
}

func TestEnsureRoleAllowed_MissingWorkflowRefusesRetiredRoles(t *testing.T) {
	retiredTempRepo(t)
	err := EnsureRoleAllowed("ghost", orchestration.RoleFeatureSpecial)
	if err == nil || !strings.Contains(err.Error(), "planner") {
		t.Fatalf("no workflow state must refuse and point at planner, got %v", err)
	}
}

func TestEnsureRoleAllowed_NonRetiredRolesAlwaysPass(t *testing.T) {
	retiredTempRepo(t)
	saveContract(t, "fresh", workflow.PlanContractUnified)
	for _, r := range []Role{
		orchestration.RolePlanner,
		orchestration.RoleSeniorEngineer,
		orchestration.RoleQASeniorEngineer,
		orchestration.RoleValidationSpec,
	} {
		if err := EnsureRoleAllowed("fresh", r); err != nil {
			t.Errorf("role %q must never be gated by the plan contract: %v", r, err)
		}
		if err := EnsureRoleAllowed("ghost", r); err != nil {
			t.Errorf("role %q must not need a workflow to be allowed: %v", r, err)
		}
	}
}
