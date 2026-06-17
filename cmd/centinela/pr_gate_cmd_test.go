package main

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

func TestPrGateFails_FailAlwaysBlocks(t *testing.T) {
	cfg := &config.Config{}
	res := []gates.Result{{Name: "G1", Status: gates.Fail}}
	if !prGateFails(cfg, res) {
		t.Fatal("a Fail result must always block the PR gate")
	}
}

func TestPrGateFails_WarnDependsOnConfig(t *testing.T) {
	res := []gates.Result{{Name: "style", Status: gates.Warn}}
	def := &config.Config{}
	if prGateFails(def, res) {
		t.Fatal("a Warn must NOT block when fail_on_warning is false")
	}
	on := &config.Config{}
	on.PrGate.FailOnWarning = true
	if !prGateFails(on, res) {
		t.Fatal("a Warn must block when fail_on_warning is true")
	}
}

func TestPrGateFails_AllPassDoesNotBlock(t *testing.T) {
	cfg := &config.Config{}
	cfg.PrGate.FailOnWarning = true
	res := []gates.Result{{Name: "G1", Status: gates.Pass}, {Name: "i18n", Status: gates.Skip}}
	if prGateFails(cfg, res) {
		t.Fatal("an all-pass/skip changeset must not block")
	}
}

// Degrade path: outside a git repo, pr-gate renders the Markdown verdict to
// stdout and (for a clean repo) exits 0 — it never crashes or posts.
func TestRunPrGate_DegradeRendersMarkdown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	out, err := pcCmd(t, "[gates]\nfile_size = false\ni18n = false\n", func(c *cobra.Command) error {
		return runPrGate(c, nil)
	})
	if err != nil {
		t.Fatalf("clean degrade must exit 0: %v", err)
	}
	if !strings.Contains(out, ui.MarkdownMarker) {
		t.Fatalf("verdict markdown must still be printed on degrade: %q", out)
	}
}

// Fail path: a G1-violating file in the (degraded full) scan makes pr-gate
// render the failing verdict and return a non-nil error (non-zero exit).
func TestRunPrGate_FailReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent")
	}
	out, err := pcCmd(t, "[gates]\nfile_size = true\ni18n = false\n", func(c *cobra.Command) error {
		if e := os.MkdirAll("internal", 0o755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile("internal/big.go", []byte(pcFiller(140)), 0o644); e != nil {
			t.Fatal(e)
		}
		return runPrGate(c, nil)
	})
	if err == nil {
		t.Fatalf("a fail-severity gate must make pr-gate exit non-zero:\n%s", out)
	}
	if !strings.Contains(out, "❌") {
		t.Fatalf("failing verdict must render the fail marker: %q", out)
	}
}
