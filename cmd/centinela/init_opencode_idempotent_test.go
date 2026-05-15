package main

import (
	"os"
	"testing"
)

func TestRunInit_OpenCode_Idempotent(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldAgent, oldLocal := agentFlag, localFlag
	defer func() { agentFlag, localFlag = oldAgent, oldLocal }()
	agentFlag, localFlag = "opencode", false

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("first runInit: %v", err)
	}
	// Second call must hit the "already configured" / no-op branches.
	if err := runInit(nil, nil); err != nil {
		t.Fatalf("second runInit must be a no-op, got: %v", err)
	}
}

func TestRunInit_Both_AgentsCovered(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldAgent, oldLocal := agentFlag, localFlag
	defer func() { agentFlag, localFlag = oldAgent, oldLocal }()
	agentFlag, localFlag = "both", false

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit both: %v", err)
	}
	for _, want := range []string{"opencode.json", ".claude/settings.json", "AGENTS.md", "CLAUDE.md"} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
}
