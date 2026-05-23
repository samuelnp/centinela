package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func wfSlice(n int) []*workflow.Workflow {
	out := make([]*workflow.Workflow, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, workflow.New("f"))
	}
	return out
}

// Spec scenario 5: above the cap returns the front `max` and the omitted count.
func TestCapActive_AboveCapReportsMore(t *testing.T) {
	shown, more := workflow.CapActive(wfSlice(7), 5)
	if len(shown) != 5 {
		t.Fatalf("expected 5 shown, got %d", len(shown))
	}
	if more != 2 {
		t.Fatalf("expected more=2, got %d", more)
	}
}

// Spec scenario 6: at-or-below the cap returns everything and more=0.
func TestCapActive_AtOrBelowCapNoMore(t *testing.T) {
	for _, n := range []int{0, 1, 3, 5} {
		shown, more := workflow.CapActive(wfSlice(n), 5)
		if len(shown) != n {
			t.Fatalf("n=%d: expected %d shown, got %d", n, n, len(shown))
		}
		if more != 0 {
			t.Fatalf("n=%d: expected more=0, got %d", n, more)
		}
	}
}

// max <= 0 means no cap: everything is shown, more is 0.
func TestCapActive_NonPositiveMaxMeansNoCap(t *testing.T) {
	shown, more := workflow.CapActive(wfSlice(4), 0)
	if len(shown) != 4 || more != 0 {
		t.Fatalf("max=0 should show all with more=0, got shown=%d more=%d", len(shown), more)
	}
	shown, more = workflow.CapActive(wfSlice(4), -1)
	if len(shown) != 4 || more != 0 {
		t.Fatalf("max<0 should show all with more=0, got shown=%d more=%d", len(shown), more)
	}
}
