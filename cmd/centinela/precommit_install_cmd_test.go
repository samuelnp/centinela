package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunPrecommitInstallUninstall(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	_, err := pcCmd(t, "[gates]\nfile_size = true\n", func(c *cobra.Command) error {
		pcInitGit(t)
		var buf bytes.Buffer
		c.SetOut(&buf)
		if e := runPrecommitInstall(c, nil); e != nil {
			return e
		}
		if !strings.Contains(buf.String(), "Installed") {
			t.Fatalf("install must confirm: %q", buf.String())
		}
		// Idempotent second install.
		buf.Reset()
		if e := runPrecommitInstall(c, nil); e != nil {
			return e
		}
		if !strings.Contains(buf.String(), "already installed") {
			t.Fatalf("second install must report no change: %q", buf.String())
		}
		buf.Reset()
		if e := runPrecommitUninstall(c, nil); e != nil {
			return e
		}
		if !strings.Contains(buf.String(), "Removed") {
			t.Fatalf("uninstall must confirm removal: %q", buf.String())
		}
		buf.Reset()
		if e := runPrecommitUninstall(c, nil); e != nil {
			return e
		}
		if !strings.Contains(buf.String(), "No centinela") {
			t.Fatalf("second uninstall must report nothing to remove: %q", buf.String())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("install/uninstall round-trip: %v", err)
	}
}

func TestHooksDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	// Inside a git repo, hooksDir resolves under .git.
	_, _ = pcCmd(t, "[gates]\nfile_size = true\n", func(c *cobra.Command) error {
		pcInitGit(t)
		if !strings.Contains(filepath.ToSlash(hooksDir()), ".git") {
			t.Fatalf("hooksDir in a repo must point under .git, got %q", hooksDir())
		}
		return nil
	})
	// Outside a repo, hooksDir falls back to .git/hooks.
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if got := filepath.ToSlash(hooksDir()); !strings.HasSuffix(got, ".git/hooks") {
		t.Fatalf("hooksDir fallback must be .git/hooks, got %q", got)
	}
}
