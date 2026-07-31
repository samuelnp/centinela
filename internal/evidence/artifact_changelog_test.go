package evidence

import (
	"slices"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestKindChangelogIsAllowed(t *testing.T) {
	if !slices.Contains(KindsAllowed(), KindChangelog) {
		t.Fatalf("changelog must be an allowed artifact kind, got %v", KindsAllowed())
	}
	got, err := ParseKind("changelog")
	if err != nil || got != KindChangelog {
		t.Fatalf("ParseKind(changelog) = %v, %v", got, err)
	}
}

func TestRenderTemplateChangelogEmitsNonBlankOneLiner(t *testing.T) {
	paths, bodies, err := RenderTemplate(KindChangelog, "right-size-docs-step")
	if err != nil {
		t.Fatalf("render changelog: %v", err)
	}
	if len(paths) != 1 || len(bodies) != 1 {
		t.Fatalf("changelog kind must emit exactly one file, got %d paths", len(paths))
	}
	if !strings.HasSuffix(paths[0], "right-size-docs-step-changelog.md") {
		t.Fatalf("changelog path mismatch: %s", paths[0])
	}
	// The stub's entry line must be non-blank so the docs gate reports it as a
	// TEMPLATE rather than as an empty file — the two failures carry different
	// remedies. The stub deliberately does NOT pass that gate; the scaffold
	// exists to be filled in, and its own gate now says so.
	first := strings.SplitN(string(bodies[0]), "\n", 2)[0]
	if strings.TrimSpace(first) == "" {
		t.Fatal("changelog stub first line must be non-blank")
	}
	if !orchestration.UnreplacedSlot(first) {
		t.Fatalf("the scaffolded changelog must still read as a template: %q", first)
	}
}
