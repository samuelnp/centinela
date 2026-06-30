package main

import (
	"os"
	"testing"
)

func TestPrewriteTargets_Branches(t *testing.T) {
	var in prewriteInput
	in.ToolInput.FilePath = "/abs/a.go"
	if got := prewriteTargets(in); len(got) != 1 || got[0] != "/abs/a.go" {
		t.Fatalf("file_path branch = %v", got)
	}
	in = prewriteInput{}
	in.ToolInput.FilePath2 = "/abs/b.go"
	if got := prewriteTargets(in); len(got) != 1 || got[0] != "/abs/b.go" {
		t.Fatalf("filePath branch = %v", got)
	}
	in = prewriteInput{}
	in.ToolInput.Command = "*** Add File: internal/c.go\n"
	if got := prewriteTargets(in); len(got) != 1 || got[0] != "internal/c.go" {
		t.Fatalf("apply_patch branch = %v", got)
	}
}

// TestRunHookPrewrite_ApplyPatchRelativeBlocks drives the REAL apply_patch path
// (no stub): a repo-relative code path with no workflow must exit 2. This is the
// cmd-level guard for the relative-path regression.
func TestRunHookPrewrite_ApplyPatchRelativeBlocks(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	exitCode := 0
	old := exitPrewrite
	defer func() { exitPrewrite = old }()
	exitPrewrite = func(c int) { exitCode = c }

	payload := `{"tool_input":{"command":"*** Begin Patch\n*** Add File: internal/foo.go\n*** End Patch"}}`
	withStdin(t, payload, func() {
		_ = runHookPrewrite(nil, nil)
	})
	if exitCode != 2 {
		t.Fatalf("relative apply_patch code path should exit 2, got %d", exitCode)
	}
}

func TestRunHookPrewrite_ApplyPatchDocsAllowed(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	exitCode := 0
	old := exitPrewrite
	defer func() { exitPrewrite = old }()
	exitPrewrite = func(c int) { exitCode = c }

	payload := `{"tool_input":{"command":"*** Add File: notes.md\n"}}`
	withStdin(t, payload, func() {
		_ = runHookPrewrite(nil, nil)
	})
	if exitCode != 0 {
		t.Fatalf("docs apply_patch path should not block, got exit %d", exitCode)
	}
}
