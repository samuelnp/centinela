package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/brownfield-setup-detection.feature

// Scenario: Brownfield repo with only a Makefile is detected as brownfield
// Scenario: Brownfield directive instructs enrich-then-confirm workflow
func TestAccBrownfield_MakefileAndEnrichConfirm(t *testing.T) {
	out, code := runSetupHook(t, setupRepo(t, "Makefile"))
	if code != 0 || !strings.Contains(out, "BROWNFIELD PROJECT DETECTED") {
		t.Fatalf("Makefile-only must be brownfield (code %d):\n%s", code, out)
	}
	// A Cargo.toml repo must carry enrich-then-confirm guidance and must NOT tell
	// the agent to ignore the user's message (that is the greenfield wording).
	cargo, _ := runSetupHook(t, setupRepo(t, "Cargo.toml"))
	if !strings.Contains(cargo, "ENRICH") || !strings.Contains(cargo, "confirm") {
		t.Fatalf("expected enrich-then-confirm guidance:\n%s", cargo)
	}
	if strings.Contains(cargo, "Do not answer the user's message") {
		t.Fatalf("brownfield must not instruct the agent to ignore the user:\n%s", cargo)
	}
}

// Scenario: Greenfield empty repo still emits the existing question-based setup directive
// Scenario: Empty src/ directory is NOT a brownfield signal
// Scenario: HasSource detector does not walk subdirectories (cheap root-only check)
func TestAccBrownfield_GreenfieldPathsStayGreenfield(t *testing.T) {
	cases := map[string][]string{
		"empty repo":    nil,                         // no source at all
		"empty src dir": {"src/"},                    // empty src is not a signal
		"nested only":   {"deep/nested/app/main.go"}, // source only deep, no root signal
	}
	for name, files := range cases {
		out, code := runSetupHook(t, setupRepo(t, files...))
		if code != 0 {
			t.Fatalf("%s: expected exit 0, got %d:\n%s", name, code, out)
		}
		if !strings.Contains(out, "CENTINELA DIRECTIVE: setup required") {
			t.Fatalf("%s: expected greenfield directive, got:\n%s", name, out)
		}
		if strings.Contains(out, "BROWNFIELD") {
			t.Fatalf("%s: must not be brownfield, got:\n%s", name, out)
		}
	}
}

// Scenario: PROJECT.md already present bypasses both setup directives
func TestAccBrownfield_ProjectMdBypassesSetup(t *testing.T) {
	out, code := runSetupHook(t, setupRepo(t, "go.mod", "PROJECT.md"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, out)
	}
	if strings.Contains(out, "BROWNFIELD") ||
		strings.Contains(out, "CENTINELA DIRECTIVE: setup required") {
		t.Fatalf("PROJECT.md present must bypass setup directives, got:\n%s", out)
	}
}
