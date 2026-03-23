package main

import (
	"os"
	"testing"
)

func TestRunInitClaudeLocal(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldAgent, oldLocal := agentFlag, localFlag
	defer func() { agentFlag, localFlag = oldAgent, oldLocal }()
	agentFlag, localFlag = "claude", true

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit claude local failed: %v", err)
	}
	if _, err := os.Stat(".claude/settings.local.json"); err != nil {
		t.Fatalf("missing local settings: %v", err)
	}
}
