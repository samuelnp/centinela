package evidence

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestOrchValidateImplFlagsMissingEvidence(t *testing.T) {
	chdirToTemp(t)
	err := orchValidateImpl(".workflow/missing.json", "alpha", "plan", orchestration.RoleBigThinker, nil)
	if err == nil {
		t.Fatal("expected error on missing file")
	}
}

func TestOrchValidateImplAcceptsValid(t *testing.T) {
	chdirToTemp(t)
	// Write a feature doc so requiredPlanInputs is satisfied.
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/features/alpha.md", []byte("# alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("docs/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/plans/alpha.md", []byte("# plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := Skeleton("alpha", orchestration.RoleBigThinker, "v1")
	s.Inputs = []string{"docs/features/alpha.md", "docs/plans/alpha.md"}
	s.Outputs = []string{"docs/plans/alpha.md"}
	s.EdgeCases = []string{"safe"}
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, s); err != nil {
		t.Fatal(err)
	}
	path := pathFor("alpha", orchestration.RoleBigThinker)
	if err := orchValidateImpl(path, "alpha", "plan", orchestration.RoleBigThinker, nil); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}
