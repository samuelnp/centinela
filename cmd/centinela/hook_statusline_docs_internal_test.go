package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// An internal docs step nags for the one-line changelog, not the portal.
func TestStatuslineRulesDocsInternalChangelog(t *testing.T) {
	d := t.TempDir()
	o := withDir(t, d)
	defer o()
	wf := workflow.New("alpha")
	wf.CurrentStep = "docs"
	cfg := &config.Config{}
	mkdir(t, ".workflow")
	// Internal feature: brief declares no user-facing surface.
	mkdir(t, "docs/features")
	write(t, "docs/features/alpha.md", "# alpha\n")
	workflow.Save(wf) //nolint:errcheck

	block, next := statusBlockAndNext(wf, cfg)
	if block != "MISSING_DOCS_OUTPUT" || next != "write-changelog" {
		t.Fatalf("internal docs step must nag for the changelog, got %s/%s", block, next)
	}

	// Once the changelog exists the docs-output nag clears (advancing past it).
	write(t, ".workflow/alpha-changelog.md", "- refactor: tidy\n")
	block, _ = statusBlockAndNext(wf, cfg)
	if block == "MISSING_DOCS_OUTPUT" {
		t.Fatalf("changelog present should clear the docs-output nag, got %s", block)
	}
}
