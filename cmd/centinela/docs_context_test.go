package main

import (
	"os"
	"strings"
	"testing"
)

func docsContextRepo(t *testing.T, withPlanAndSpec bool) {
	t.Helper()
	chdir(t, t.TempDir())
	os.MkdirAll("docs/features", 0755)                        //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("# f\n"), 0644) //nolint:errcheck
	if !withPlanAndSpec {
		return
	}
	os.MkdirAll("docs/plans", 0755)                             //nolint:errcheck
	os.MkdirAll("specs", 0755)                                  //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("# plan\n"), 0644)   //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: f"), 0644) //nolint:errcheck
}

func TestRunDocsContextHappyPathPrintsMarkdown(t *testing.T) {
	docsContextRepo(t, true)
	var err error
	out := captureStdout(t, func() { err = runDocsContext(nil, []string{"f"}) })
	if err != nil {
		t.Fatalf("docs context should pass: %v", err)
	}
	for _, want := range []string{"# Docs Context: f", "## Feature brief", "## Plan", "## Spec scenarios", "## Changelog draft"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestRunDocsContextRejectsBadSlug(t *testing.T) {
	docsContextRepo(t, true)
	for _, slug := range []string{"../escape", "UPPER", "spaced slug"} {
		if err := runDocsContext(nil, []string{slug}); err == nil {
			t.Fatalf("slug %q must be rejected before any path is built", slug)
		}
	}
}

func TestRunDocsContextMissingInputsAggregated(t *testing.T) {
	docsContextRepo(t, false)
	err := runDocsContext(nil, []string{"f"})
	if err == nil {
		t.Fatal("missing plan+spec must fail")
	}
	for _, want := range []string{"docs/plans/f.md", "specs/f.feature"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error must name %s, got %v", want, err)
		}
	}
}

// Deleted subcommands must fail loudly: the docs parent RunE rejects unknown
// subcommands with a non-nil error (RunE exit 1), never help with exit 0.
func TestDocsCmdUnknownSubcommandsError(t *testing.T) {
	for _, sub := range []string{"generate", "validate"} {
		err := docsCmd.RunE(docsCmd, []string{sub})
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("docs %s must fail as unknown command, got %v", sub, err)
		}
	}
}
