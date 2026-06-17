package insights

import (
	"reflect"
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Gates must be a pure pass-through to the unexported gates ranking, so the plan
// advisor can never diverge from `centinela insights` (count desc, key asc,
// empty Gate → "<none>").
func TestGatesExportedEqualsUnexported(t *testing.T) {
	ev := []telemetry.Event{
		gf("coverage"), gf("coverage"), gf("coverage"),
		gf("import-graph"), gf("import-graph"),
		gf("g1-file-size"),
		gf(""),
		{Type: telemetry.TypeBlock},
	}
	for _, n := range []int{1, 2, 5, 10} {
		if got, want := Gates(ev, n), gates(ev, n); !reflect.DeepEqual(got, want) {
			t.Fatalf("Gates(ev,%d)=%+v, gates=%+v", n, got, want)
		}
	}
}

// The exported wrapper inherits the <none> bucket for an empty Gate field.
func TestGatesExportedBucketsEmptyGateAsNone(t *testing.T) {
	got := Gates([]telemetry.Event{gf("")}, 5)
	if len(got) != 1 || got[0].Key != "<none>" {
		t.Fatalf("Gates empty-gate = %+v, want [<none>]", got)
	}
}

// The exported wrapper agrees with Compute().Gates on the same input.
func TestGatesExportedMatchesCompute(t *testing.T) {
	ev := []telemetry.Event{gf("z"), gf("a"), gf("a"), gf("m")}
	if got, want := Gates(ev, 5), Compute(ev, 5).Gates; !reflect.DeepEqual(got, want) {
		t.Fatalf("Gates=%+v, Compute().Gates=%+v", got, want)
	}
}
