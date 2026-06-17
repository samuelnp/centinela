package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// pcCmd runs fn with CWD set to a fresh temp dir holding a centinela.toml, and
// returns the combined stdout+stderr the cobra command wrote plus fn's error.
func pcCmd(t *testing.T, toml string, fn func(*cobra.Command) error) (string, error) {
	t.Helper()
	d := t.TempDir()
	if r, err := filepath.EvalSymlinks(d); err == nil {
		d = r
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.WriteFile(filepath.Join(d, "centinela.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	c := &cobra.Command{}
	c.SetOut(&buf)
	c.SetErr(&buf)
	err := fn(c)
	return buf.String(), err
}

func pcInitGit(t *testing.T) {
	t.Helper()
	for _, a := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "qa@x.dev"},
		{"config", "user.name", "QA"},
	} {
		if out, err := exec.Command("git", a...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", a, err, out)
		}
	}
}

func pcFiller(n int) string {
	return "package x\n" + strings.Repeat("// filler\n", n)
}

// Degrade path: outside a git repo, precommit must NOT block (nil err).
func TestRunPrecommit_DegradeNeverBlocks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	out, err := pcCmd(t, "[gates]\nfile_size = true\n", func(c *cobra.Command) error {
		return runPrecommit(c, nil)
	})
	if err != nil {
		t.Fatalf("not-a-repo must degrade, not block: %v", err)
	}
	if !strings.Contains(strings.ToLower(out), "nothing to gate") {
		t.Fatalf("degrade must print a notice: %q", out)
	}
}

// Fail path: a staged oversized file blocks the commit with a non-nil error.
func TestRunPrecommit_StagedFailBlocks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	out, err := pcCmd(t, "[gates]\nfile_size = true\ni18n = false\n", func(c *cobra.Command) error {
		pcInitGit(t)
		if e := os.MkdirAll("internal", 0o755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile("internal/big.go", []byte(pcFiller(140)), 0o644); e != nil {
			t.Fatal(e)
		}
		if o, e := exec.Command("git", "add", "internal/big.go").CombinedOutput(); e != nil {
			t.Fatalf("git add: %v\n%s", e, o)
		}
		return runPrecommit(c, nil)
	})
	if err == nil {
		t.Fatalf("oversized staged file must block:\n%s", out)
	}
	if !strings.Contains(out, "File Size") {
		t.Fatalf("the failing gate must be rendered: %q", out)
	}
}
