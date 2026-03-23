package main

import (
	"os"
	"testing"
)

func TestSetupOpenCodeAlreadyConfigured(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("opencode.json", []byte(`{"$schema":"https://opencode.ai/config.json","instructions":["CLAUDE.md"]}`), 0644) //nolint:errcheck
	os.MkdirAll(".opencode/plugins", 0755)                                                                                    //nolint:errcheck
	os.WriteFile(".opencode/plugins/centinela.js", []byte("x"), 0644)                                                         //nolint:errcheck
	os.WriteFile("AGENTS.md", []byte("x"), 0644)                                                                              //nolint:errcheck
	if err := setupOpenCode(); err != nil {
		t.Fatalf("setupOpenCode already configured should pass: %v", err)
	}
}
