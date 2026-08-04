package roadmap

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// The knob governs VCS policy only: regeneration is correctness and still runs.
func TestSyncCommitDisabledStillRegenerates(t *testing.T) {
	syncRepo(t, syncBody)
	c := &fakeCommitter{}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: false, C: c})
	if !rep.Regenerated {
		t.Fatal("regeneration must happen even with auto-commit disabled")
	}
	if c.calls != 0 || rep.Committed || rep.Warn {
		t.Fatalf("no commit may be attempted: calls=%d rep=%+v", c.calls, rep)
	}
	if !strings.Contains(rep.Reason, "auto-commit is disabled") {
		t.Fatalf("reason = %q", rep.Reason)
	}
	if _, err := os.Stat("ROADMAP.md"); err != nil {
		t.Fatalf("ROADMAP.md must exist: %v", err)
	}
}

func TestSyncCommitterErrorBecomesAWarning(t *testing.T) {
	syncRepo(t, syncBody)
	c := &fakeCommitter{err: errors.New("merge in progress")}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: true, C: c})
	if rep.Committed || !rep.Warn || rep.Reason != "merge in progress" {
		t.Fatalf("report = %+v", rep)
	}
}

func TestSyncNoChangeIsNotAWarning(t *testing.T) {
	syncRepo(t, syncBody)
	c := &fakeCommitter{err: ErrNoChange}
	rep := Sync(SyncOptions{Verb: "defer", Subject: "x", Commit: true, C: c})
	if rep.Committed || rep.Warn || rep.Reason != ErrNoChange.Error() {
		t.Fatalf("report = %+v", rep)
	}
}
