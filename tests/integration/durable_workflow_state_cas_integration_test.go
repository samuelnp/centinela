package integration_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/samuelnp/centinela/internal/doctor"
	"github.com/samuelnp/centinela/internal/workflow"
)

// The honest, measured form of the CAS guarantee: it detects STALENESS. With
// competing writers whose saves do not overlap — the realistic multi-process
// shape, since every command marshals and renames in well under the time it
// spends working — detection is total: one writer wins, every other is refused,
// and nothing is silently lost. (Fully OVERLAPPING saves are NOT protected; see
// checkNotStale and the close-state-save-toctou deferral.)
func TestSerialisedConcurrentWritersLoseNothing(t *testing.T) {
	dwsProject(t)
	const writers = 12

	loaded := make([]*workflow.Workflow, writers)
	var wg sync.WaitGroup
	for i := range loaded {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			wf, err := workflow.Load("alpha")
			if err != nil {
				return
			}
			wf.SetModelRoute(fmt.Sprintf("role-%d", i), workflow.ModelRoute{Tier: "balanced"})
			loaded[i] = wf
		}(i)
	}
	wg.Wait()

	// The saves are serialised by construction — this loop IS the serialisation,
	// which is the shape a real multi-process race takes (each command marshals
	// and renames in milliseconds; the reads are what overlap for minutes).
	ok, refused := 0, 0
	for _, wf := range loaded {
		if wf == nil {
			t.Fatal("every concurrent Load must succeed")
		}
		if err := workflow.Save(wf); err != nil {
			refused++
		} else {
			ok++
		}
	}
	if ok != 1 || refused != writers-1 {
		t.Fatalf("ok=%d refused=%d, want 1 and %d", ok, refused, writers-1)
	}
	final, err := workflow.Load("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(final.ModelRoutes) != ok {
		t.Fatalf("%d routes on disk, want exactly the %d that reported success — "+
			"any other count means a reported success was silently lost",
			len(final.ModelRoutes), ok)
	}
}

// DELETE THIS TEST when the `doctor-detect-corrupt-state-file` deferral lands.
//
// It is the executable statement of a known gap, not an endorsement: a state
// file corrupted by the PRE-atomic writer is invisible to doctor, which prints
// a green "no orphaned .workflow state" while the only trace is a check-less
// stderr warning. This feature stops NEW corruption; it neither detects nor
// repairs what the old writer already broke. When someone teaches doctor to
// report it, this test fails and points at the deferral it belongs to.
func TestDoctorIsStillBlindToACorruptStateFile(t *testing.T) {
	dir := dwsBareProject(t)
	dwsWriteRaw(t, "broken", `{"feature":"broken","currentStep":"cod`)

	got := dwsDiagnose(t, dir)
	ws, ok := got["workflow-state"]
	if !ok {
		t.Fatal("doctor ran no workflow-state check")
	}
	if ws.Status != doctor.OK {
		t.Fatalf("KNOWN GAP CLOSED — doctor now reports a corrupt state file (%+v). "+
			"Delete this test and close doctor-detect-corrupt-state-file.", ws)
	}
	for name, d := range got {
		for _, detail := range d.Details {
			if strings.Contains(detail, "broken") {
				t.Fatalf("KNOWN GAP CLOSED — check %q now names it: %s. Delete this "+
					"test and close doctor-detect-corrupt-state-file.", name, detail)
			}
		}
	}
}
