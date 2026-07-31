package docsctx

import (
	"testing"
)

func TestLoadChangelogOptional(t *testing.T) {
	files := allInputs()
	delete(files, ".workflow/f-changelog.md")
	seed(t, files)
	ctx, err := Load("f")
	if err != nil {
		t.Fatalf("absent changelog must not fail Load: %v", err)
	}
	if ctx.Changelog.Present || ctx.Changelog.Source != ".workflow/f-changelog.md" {
		t.Fatalf("absent changelog must stay addressable: %+v", ctx.Changelog)
	}
}

// Boundary pin: an empty-but-present (0-byte) required input is NOT missing —
// Load succeeds and the section renders empty. Documented current behavior.
func TestLoadEmptyButPresentInputPasses(t *testing.T) {
	files := allInputs()
	files["docs/features/f.md"] = ""
	seed(t, files)
	ctx, err := Load("f")
	if err != nil {
		t.Fatalf("0-byte brief must count as present: %v", err)
	}
	if !ctx.Brief.Present || ctx.Brief.Body != "" {
		t.Fatalf("0-byte brief must be Present with empty body: %+v", ctx.Brief)
	}
}
