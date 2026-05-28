package evidence

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestSchemaSkeletonReturnsPlaceholder(t *testing.T) {
	out, err := SchemaSkeleton(orchestration.RoleFeatureSpecial, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<feature-slug>") {
		t.Fatalf("skeleton missing placeholder: %s", out)
	}
}

func TestReadFieldEveryScalar(t *testing.T) {
	s := Skeleton("alpha", orchestration.RoleUXUISpecialist, "v1")
	s.Inputs = []string{"x"}
	s.Outputs = []string{"y"}
	s.EdgeCases = []string{"z"}
	fields := []string{"feature", "step", "role", "status", "generatedAt", "handoffTo", "mobileFirst", "_meta"}
	for _, f := range fields {
		if _, err := ReadField(s, f); err != nil {
			t.Errorf("ReadField(%s): %v", f, err)
		}
	}
}

func TestRepairIgnoresMissingPaths(t *testing.T) {
	chdirToTemp(t)
	removed, err := Repair("ghost")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removals, got %v", removed)
	}
}
