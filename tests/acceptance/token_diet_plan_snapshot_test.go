// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// tdWriteBriefs drops n empty brief files under docs/features/ in the
// current (chdir'ed) directory.
func tdWriteBriefs(t *testing.T, n int) {
	t.Helper()
	if err := os.MkdirAll("docs/features", 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := filepath.Join("docs", "features", fmt.Sprintf("brief-%03d.md", i))
		if err := os.WriteFile(p, []byte("# brief\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Scenario: The required plan-input set does not grow with the number of briefs
func TestTD_RequiredPlanInputsDoesNotGrowWithBriefCount(t *testing.T) {
	t.Chdir(t.TempDir())
	tdWriteBriefs(t, 120)

	got := orchestration.RequiredPlanInputs("token-diet")
	want := []string{"docs/features/token-diet.md", "docs/plans/token-diet.md"}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want exactly %v", got, want)
	}
	for _, entry := range got {
		if strings.Contains(entry, "brief-") {
			t.Fatalf("no other feature's brief may appear in the set: %v", got)
		}
	}
}

// Scenario: The required set is identical in an empty repository
func TestTD_RequiredPlanInputsIdenticalEmptyVsPopulated(t *testing.T) {
	t.Chdir(t.TempDir())
	empty := orchestration.RequiredPlanInputs("token-diet")

	t.Chdir(t.TempDir())
	tdWriteBriefs(t, 120)
	populated := orchestration.RequiredPlanInputs("token-diet")

	if !reflect.DeepEqual(empty, populated) {
		t.Fatalf("empty-repo set %v must equal 120-brief set %v", empty, populated)
	}
}

// Scenario: The required set is derived by construction, not by existence
func TestTD_RequiredPlanInputsDerivedByConstruction(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, p := range []string{"docs/features/ghost.md", "docs/plans/ghost.md"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("%s must not exist for this test", p)
		}
	}
	got := orchestration.RequiredPlanInputs("ghost")
	want := []string{"docs/features/ghost.md", "docs/plans/ghost.md"}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("both paths should still be required by construction: got %v", got)
	}
}
