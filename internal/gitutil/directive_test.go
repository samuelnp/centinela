package gitutil

import (
	"strings"
	"testing"
)

// TestDeliveryDirectiveWithOptions emits the 2-line directive listing only the
// offered options as exact deliver commands.
func TestDeliveryDirectiveWithOptions(t *testing.T) {
	out := DeliveryDirective("alpha", []Option{OptionPR, OptionMerge})
	for _, want := range []string{
		"CENTINELA DIRECTIVE:", "alpha", "ask the user how to deliver",
		"do NOT push or merge", "centinela deliver alpha --via pr",
		"centinela deliver alpha --via merge",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("directive missing %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("expected a 2-line directive, got:\n%s", out)
	}
}

// TestDeliveryDirectivePROnly lists only the PR command when merge is not offered.
func TestDeliveryDirectivePROnly(t *testing.T) {
	out := DeliveryDirective("beta", []Option{OptionPR})
	if !strings.Contains(out, "--via pr") || strings.Contains(out, "--via merge") {
		t.Fatalf("pr-only directive wrong:\n%s", out)
	}
}

// TestDeliveryDirectiveEmpty collapses to a single no-target line.
func TestDeliveryDirectiveEmpty(t *testing.T) {
	out := DeliveryDirective("gamma", nil)
	if strings.Contains(out, "\n") {
		t.Fatalf("empty directive should be one line:\n%s", out)
	}
	if !strings.Contains(out, "no delivery target") || !strings.Contains(out, "gamma") {
		t.Fatalf("empty directive missing no-target message:\n%s", out)
	}
}

// TestGitHubCLIAvailable just exercises the lookup (result depends on PATH).
func TestGitHubCLIAvailable(t *testing.T) {
	_ = GitHubCLIAvailable()
}
