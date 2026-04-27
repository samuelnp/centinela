package orchestration

import (
	"strings"
	"testing"
)

func TestValidateUXEvidenceRequiresMobileFirstAndTags(t *testing.T) {
	if err := validateUXEvidence("x", RoleUXUISpecialist, requiredUXTags, nil); err == nil || !strings.Contains(err.Error(), "mobileFirst") {
		t.Fatalf("expected mobileFirst error, got %v", err)
	}
	mobileFirst := true
	if err := validateUXEvidence("x", RoleUXUISpecialist, []string{"mobile-first"}, &mobileFirst); err == nil || !strings.Contains(err.Error(), "visual-hierarchy") {
		t.Fatalf("expected missing ux tag error, got %v", err)
	}
	if err := validateUXEvidence("x", RoleUXUISpecialist, requiredUXTags, &mobileFirst); err != nil {
		t.Fatalf("expected ux evidence success, got %v", err)
	}
	if err := validateUXEvidence("x", RoleSeniorEngineer, nil, nil); err != nil {
		t.Fatalf("expected non-ux bypass, got %v", err)
	}
}
