// Acceptance: specs/roadmap-state-hygiene.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// Scenario: two concurrent mutations in one checkout both survive
//
// Two real `centinela roadmap defer` PROCESSES, started together in one
// checkout. Before the roadmap-state lock this destroyed one deferral in 5 runs
// out of 5: both read the same document, the second os.Rename won, and the
// winner's auto-commit left a CLEAN tree — so `git status`, the one signal that
// existed before this feature, showed nothing at all.
func TestRsh_ConcurrentDefersBothSurvive(t *testing.T) {
	bin := buildCent(t)
	dir := rshRepo(t, rshBaseRoadmap)
	slugs := []string{"race-one", "race-two", "race-three"}

	var wg sync.WaitGroup
	outs := make([]string, len(slugs))
	for i, slug := range slugs {
		wg.Add(1)
		go func(i int, slug string) {
			defer wg.Done()
			c := exec.Command(bin, "roadmap", "defer", slug, "--summary", "concurrent")
			c.Dir = dir
			out, err := c.CombinedOutput()
			outs[i] = string(out)
			if err != nil {
				t.Errorf("worker %s: %v\n%s", slug, err, out)
			}
		}(i, slug)
	}
	wg.Wait()

	body := mustRead(t, dir+"/.workflow/roadmap.json")
	for i, slug := range slugs {
		if !strings.Contains(body, `"`+slug+`"`) {
			t.Fatalf("%q was destroyed by a concurrent mutation\nworker said:\n%s\nroadmap.json:\n%s",
				slug, outs[i], body)
		}
	}
	// Scenario: a lost record is never reported as merely uncommitted
	// No worker may claim the tree holds something that is not there. The
	// read-back that makes this true is unit-tested in internal/roadmap
	// (TestSyncVerifiesTheStateIsOnDisk) and internal/ui
	// (TestRenderRoadmapSyncRefusesToClaimAnUnverifiedTree); here it is
	// asserted as an invariant over three real racing processes.
	for i, out := range outs {
		if strings.Contains(out, "in your working tree") && !strings.Contains(body, slugs[i]) {
			t.Fatalf("worker %d claimed the tree holds a record that is gone:\n%s", i, out)
		}
	}
	// The advisory lock must never become an untracked file in the tree: an
	// aborted ff-merge on a dirty tree is what started this feature.
	if status := rshGitOut(t, dir, "status", "--porcelain"); status != "" {
		t.Fatalf("the tree must be clean after concurrent mutations, got:\n%s", status)
	}
}
