package main

import (
	"os"
	"testing"
)

func TestRunInitInvalidAgent(t *testing.T) {
	old := agentFlag
	defer func() { agentFlag = old }()
	agentFlag = "bad"
	if err := runInit(nil, nil); err == nil {
		t.Fatal("expected invalid agent error")
	}
}

func TestRunInitOpenCodeOnly(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	oldAgent, oldLocal := agentFlag, localFlag
	defer func() { agentFlag, localFlag = oldAgent, oldLocal }()
	agentFlag, localFlag = "opencode", false

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit opencode: %v", err)
	}
	if _, err := os.Stat("opencode.json"); err != nil {
		t.Fatalf("missing opencode.json: %v", err)
	}
	if _, err := os.Stat(".opencode/plugins/centinela.js"); err != nil {
		t.Fatalf("missing OpenCode plugin: %v", err)
	}
}
