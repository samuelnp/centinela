package orchestration

import "testing"

func TestRequiredRolesUnknownStep(t *testing.T) {
	if got := RequiredRoles("tests"); len(got) != 1 || got[0] != RoleQASeniorEngineer {
		t.Fatalf("expected qa role for tests step, got %v", got)
	}
	if got := RequiredRoles("docs"); len(got) != 1 || got[0] != RoleDocsSpecialist {
		t.Fatalf("expected docs role for docs step, got %v", got)
	}
	if got := RequiredRoles("unknown"); got != nil {
		t.Fatalf("expected nil roles for unknown step, got %v", got)
	}
}
