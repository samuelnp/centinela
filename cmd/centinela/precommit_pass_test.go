package main

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

// Pass path: a clean staged file passes precommit (nil error).
func TestRunPrecommit_CleanPasses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	_, err := pcCmd(t, "[gates]\nfile_size = true\ni18n = false\n", func(c *cobra.Command) error {
		pcInitGit(t)
		if e := os.MkdirAll("internal", 0o755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile("internal/ok.go", []byte(pcFiller(10)), 0o644); e != nil {
			t.Fatal(e)
		}
		if o, e := exec.Command("git", "add", "internal/ok.go").CombinedOutput(); e != nil {
			t.Fatalf("git add: %v\n%s", e, o)
		}
		return runPrecommit(c, nil)
	})
	if err != nil {
		t.Fatalf("clean staged file must pass: %v", err)
	}
}
