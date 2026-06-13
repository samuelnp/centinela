package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// start --profile pins the chosen profile onto the persisted workflow, and the
// pinned profile decides the orchestration mode (outcome → no subagent evidence).
func TestRunStart_ProfileFlagPersisted(t *testing.T) {
	old := startProfile
	defer func() { startProfile = old }()
	startProfile = config.ProfileOutcome

	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
	if err := runStart(nil, []string{"feat"}); err != nil {
		t.Fatalf("runStart: %v", err)
	}
	wf, err := workflow.Load("feat")
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}
	if wf.EnforcementProfile != config.ProfileOutcome {
		t.Fatalf("pinned profile = %q, want outcome", wf.EnforcementProfile)
	}
	if wf.OrchestrationMode != "" {
		t.Fatalf("outcome must leave orchestration mode empty, got %q", wf.OrchestrationMode)
	}
}

// With no --profile, start honors the global config profile — but does NOT pin
// it. An explicit global is left unpinned so it resolves live through
// EffectiveProfile and stays distinguishable from a per-feature pin in status
// provenance ("global" vs "--profile"). The live resolution still yields guided.
func TestRunStart_InheritsGlobalProfile(t *testing.T) {
	old := startProfile
	defer func() { startProfile = old }()
	startProfile = ""

	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: existing\n"), 0644)                       //nolint:errcheck
	os.WriteFile(config.Filename, []byte("[workflow]\nenforcement_profile=\"guided\"\n"), 0644) //nolint:errcheck
	if err := runStart(nil, []string{"feat2"}); err != nil {
		t.Fatalf("runStart: %v", err)
	}
	wf, _ := workflow.Load("feat2")
	if wf.EnforcementProfile != "" {
		t.Fatalf("explicit global must not be pinned, got %q", wf.EnforcementProfile)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got := workflow.EffectiveProfile(wf, cfg); got != config.ProfileGuided {
		t.Fatalf("effective profile should resolve global guided, got %q", got)
	}
}
