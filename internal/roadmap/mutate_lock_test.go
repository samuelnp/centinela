package roadmap

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// F1 regression, in-process tier: N mutations racing on ONE roadmap.json must
// all survive. flock is held per open file description, so separate Acquire
// calls exclude each other even inside a single process — the same property
// internal/evidence's lock test relies on.
func TestConcurrentDefersAllSurvive(t *testing.T) {
	dir := syncRepo(t, `{"phases":[{"name":"Phase 1","features":[]}]}`)
	_ = dir
	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // release all goroutines at once to maximize the overlap
			errs[i] = Defer(RoadmapFile, DeferOptions{
				Slug:    slugFor(i),
				Summary: "concurrent finding",
				Now:     time.Date(2026, 8, 4, 12, 0, i, 0, time.UTC),
			})
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("defer %d failed: %v", i, err)
		}
	}
	body, err := os.ReadFile(RoadmapFile)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		if !strings.Contains(string(body), `"`+slugFor(i)+`"`) {
			t.Fatalf("finding %q was clobbered by a concurrent mutation:\n%s", slugFor(i), body)
		}
	}
	r, err := Load()
	if err != nil {
		t.Fatalf("the merged document must still parse: %v", err)
	}
	if got := len(BacklogFeatures(r)); got != n {
		t.Fatalf("Backlog holds %d findings, want %d", got, n)
	}
}

func slugFor(i int) string {
	return "race-" + string(rune('a'+i))
}

// The lock excludes: a second acquire while one is held must time out rather
// than proceed, and the mutation that timed out must have written nothing.
func TestLockRoadmapStateExcludesAndFailsClosed(t *testing.T) {
	syncRepo(t, `{"phases":[{"name":"Phase 1","features":[]}]}`)
	before, err := os.ReadFile(RoadmapFile)
	if err != nil {
		t.Fatal(err)
	}
	release, err := lockRoadmapState(RoadmapFile)
	if err != nil {
		t.Fatal(err)
	}
	deferErr := Defer(RoadmapFile, DeferOptions{Slug: "blocked", Summary: "x"})
	release()
	if deferErr == nil {
		t.Fatal("a mutation must not proceed while another holds the lock")
	}
	if !strings.Contains(deferErr.Error(), "nothing was written") {
		t.Fatalf("the error must say the mutation did not happen: %v", deferErr)
	}
	after, err := os.ReadFile(RoadmapFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("a lock timeout must leave roadmap.json byte-identical")
	}
}
