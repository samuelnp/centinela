package evidence

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// TestSkeletonInputsStayEmpty proves the pre-fill lives in the init wiring, not
// in Skeleton — so repair (SchemaSkeleton) and the docs template are never
// poisoned with plan-snapshot inputs.
func TestSkeletonInputsStayEmpty(t *testing.T) {
	skel := Skeleton("demo", orchestration.RoleBigThinker, "1.0.0")
	if len(skel.Inputs) != 0 {
		t.Fatalf("Skeleton inputs should be empty, got %v", skel.Inputs)
	}
}

func TestSchemaSkeletonRepairInputsEmpty(t *testing.T) {
	data, err := SchemaSkeleton(orchestration.RoleBigThinker, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "docs/plans/") {
		t.Fatalf("repair skeleton must not pre-fill plan inputs: %s", data)
	}
}

func TestDocsSpecialistPairInputsEmpty(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	_, bodies, err := docsSpecialistPair("demo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(bodies[0]), "docs/plans/") {
		t.Fatalf("docs template JSON must not pre-fill plan inputs: %s", bodies[0])
	}
}
