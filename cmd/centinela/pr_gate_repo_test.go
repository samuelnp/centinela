package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
)

// Real-repo (non-degrade) path: pr-gate resolves the changed-since-base set and
// renders the verdict. A feature branch with an oversized file fails the gate.
func TestRunPrGate_RealRepoFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	gitDo := func(args ...string) {
		if o, e := exec.Command("git", args...).CombinedOutput(); e != nil {
			t.Fatalf("git %v: %v\n%s", args, e, o)
		}
	}
	out, err := pcCmd(t, "[gates]\nfile_size = true\ni18n = false\n", func(c *cobra.Command) error {
		pcInitGit(t)
		if e := os.WriteFile("README.md", []byte("seed\n"), 0o644); e != nil {
			t.Fatal(e)
		}
		gitDo("add", "README.md")
		gitDo("commit", "-q", "-m", "seed")
		gitDo("checkout", "-q", "-b", "feature")
		if e := os.MkdirAll("internal", 0o755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile("internal/big.go", []byte(pcFiller(140)), 0o644); e != nil {
			t.Fatal(e)
		}
		gitDo("add", "internal/big.go")
		gitDo("commit", "-q", "-m", "change")
		return runPrGate(c, nil)
	})
	if err == nil {
		t.Fatalf("changed oversized file must fail pr-gate:\n%s", out)
	}
	if !strings.Contains(out, ui.MarkdownMarker) || !strings.Contains(out, "❌") {
		t.Fatalf("non-degrade verdict must render markdown + fail marker: %q", out)
	}
}
