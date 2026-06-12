package evidence

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func chdirPlanTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/demo.md", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPlanInputsDelegatesForPlanRoles(t *testing.T) {
	chdirPlanTemp(t)
	want := orchestration.RequiredPlanInputs("demo")
	for _, role := range []Role{orchestration.RoleBigThinker, orchestration.RoleFeatureSpecial} {
		got := PlanInputs("demo", role)
		if len(got) != len(want) {
			t.Fatalf("%s: length mismatch got %v want %v", role, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: %q != %q", role, got[i], want[i])
			}
		}
	}
}

func TestPlanInputsNilForNonPlanRoles(t *testing.T) {
	chdirPlanTemp(t)
	planRoles := map[Role]bool{
		orchestration.RoleBigThinker:     true,
		orchestration.RoleFeatureSpecial: true,
	}
	for _, role := range AllRoles() {
		got := PlanInputs("demo", role)
		if planRoles[role] {
			if got == nil {
				t.Fatalf("plan role %s should pre-fill, got nil", role)
			}
			continue
		}
		if got != nil {
			t.Fatalf("non-plan role %s should be nil, got %v", role, got)
		}
	}
}
