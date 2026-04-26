package orchestration

import (
	"os"
	"strings"
	"testing"
)

func TestValidateActionableOutputsRoleRules(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                               //nolint:errcheck
	os.Chdir(d)                                                                     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                  //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                 //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                      //nolint:errcheck
	os.MkdirAll("internal/demo", 0755)                                              //nolint:errcheck
	os.MkdirAll("tests/unit", 0755)                                                 //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)                              //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: f"), 0644)                     //nolint:errcheck
	os.WriteFile("internal/demo/file.go", []byte("package demo"), 0644)             //nolint:errcheck
	os.MkdirAll("src/ui", 0755)                                                     //nolint:errcheck
	os.WriteFile("src/ui/page.tsx", []byte("export const Page = () => null"), 0644) //nolint:errcheck
	os.WriteFile("tests/unit/f_test.go", []byte("package unit_test"), 0644)         //nolint:errcheck
	os.WriteFile(".workflow/f-edge-cases.md", []byte("ok"), 0644)                   //nolint:errcheck
	os.WriteFile(".workflow/f-senior-engineer.md", []byte("ok"), 0644)              //nolint:errcheck
	if err := validateActionableOutputs("x", "f", RoleBigThinker, []string{"summary only"}, nil); err == nil || !strings.Contains(err.Error(), "real files") {
		t.Fatalf("expected missing file error, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleBigThinker, []string{"docs/plans/f.md"}, nil); err != nil {
		t.Fatalf("expected big-thinker success, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleSeniorEngineer, []string{".workflow/f-senior-engineer.md"}, nil); err == nil || !strings.Contains(err.Error(), "non-evidence implementation") {
		t.Fatalf("expected implementation output error, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleSeniorEngineer, []string{"internal/demo/file.go"}, nil); err != nil {
		t.Fatalf("expected senior-engineer success, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleQASeniorEngineer, []string{"tests/unit/f_test.go"}, nil); err == nil || !strings.Contains(err.Error(), "edge-cases") {
		t.Fatalf("expected qa edge-case error, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleQASeniorEngineer, []string{"tests/unit/f_test.go", ".workflow/f-edge-cases.md"}, nil); err != nil {
		t.Fatalf("expected qa success, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleDocsSpecialist, []string{"summary only"}, nil); err != nil {
		t.Fatalf("expected docs specialist bypass, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleUXUISpecialist, []string{"internal/demo/file.go"}, []string{"src/ui"}); err == nil || !strings.Contains(err.Error(), "real UI file") {
		t.Fatalf("expected ux-ui-specialist ui path error, got %v", err)
	}
	if err := validateActionableOutputs("x", "f", RoleUXUISpecialist, []string{"src/ui/page.tsx"}, []string{"src/ui"}); err != nil {
		t.Fatalf("expected ux-ui-specialist success, got %v", err)
	}
}
