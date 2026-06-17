package acceptance_test

// Acceptance: specs/governed-project-memory.feature (SC-01 through SC-13)
//
// SC-01: Edge-case lessons captured when tests step completes.
// SC-02: Gatekeeper verdict captured when validate step completes.
// SC-03: Each decision bullet becomes a separate entry when plan step completes.
// SC-04: No decision entries created when brief has no Decisions section.
// SC-05: Capture is idempotent — re-completing does not duplicate.
// SC-06: Missing artifact does not block step completion.
// SC-07: Malformed artifact does not block step completion.
// SC-08: Relevant memory recalled into plan step context.
// SC-09: Deterministic ranking — dep > shared tag > recency.
// SC-10: Empty ledger produces no error.
// SC-11: Recall caps the injected slice by count and bytes.
// SC-12: Memory disabled makes capture and recall no-ops.
// SC-13: Concurrent worktree completes do not clobber each other.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/memory"
)

func acptCfg() *config.Config {
	enabled := true
	return &config.Config{Memory: config.MemoryConfig{Enabled: &enabled, RecallMaxEntries: 10, RecallMaxBytes: 4096}}
}

func countEntries(t *testing.T) int {
	t.Helper()
	p := filepath.Join(".workflow", "memory", "entries")
	es, err := os.ReadDir(p)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range es {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			n++
		}
	}
	return n
}

// SC-01: lesson entry captured for tests step.
func TestSC01_LessonCapturedOnTestsComplete(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                              //nolint:errcheck
	os.Chdir(dir)                                                                     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                    //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- timeout lesson\n"), 0644) //nolint:errcheck

	memory.Capture("alpha", "tests", acptCfg())

	if countEntries(t) != 1 {
		t.Fatalf("SC-01 FAIL: expected 1 lesson entry, got %d", countEntries(t))
	}
}

// SC-02: verdict entry captured for validate step.
func TestSC02_VerdictCapturedOnValidateComplete(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                                              //nolint:errcheck
	os.Chdir(dir)                                                                                     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                    //nolint:errcheck
	os.WriteFile(".workflow/alpha-gatekeeper.md", []byte("Status: SAFE\nAll checks passed.\n"), 0644) //nolint:errcheck

	memory.Capture("alpha", "validate", acptCfg())

	if countEntries(t) != 1 {
		t.Fatalf("SC-02 FAIL: expected 1 verdict entry, got %d", countEntries(t))
	}
}
