package main

import (
	"os/exec"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestCov2GitOwnerResolvesAuthor drives gitOwner's success path: a resolvable
// ref returns its latest commit author.
func TestCov2GitOwnerResolvesAuthor(t *testing.T) {
	d := t.TempDir()
	git := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = d
		c.Env = append(c.Environ(),
			"GIT_AUTHOR_NAME=Ada Lovelace", "GIT_AUTHOR_EMAIL=ada@x.io",
			"GIT_COMMITTER_NAME=Ada Lovelace", "GIT_COMMITTER_EMAIL=ada@x.io")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	git("init")
	git("commit", "--allow-empty", "-m", "seed")
	git("branch", "feat")
	if owner := gitOwner(d, "feat"); owner != "Ada Lovelace" {
		t.Fatalf("expected the commit author, got %q", owner)
	}
}

// TestCov2EmitCostWarningSurvivesTelemetryReadError exercises emitCostWarning's
// ReadDefault error branch: cost active, a workflow present, but the telemetry
// log is unreadable. The warning helper must silently bail (never panic).
func TestCov2EmitCostWarningSurvivesTelemetryReadError(t *testing.T) {
	d := seedCostRepo(t)
	writeOversizeTelemetry(t, d)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	emitCostWarning(cfg) // must not panic; returns on the read error
}
