package migration

import (
	"strings"
	"testing"
)

func TestReplaceKeepBlocksKeepsTemplateBodyForUnknownID(t *testing.T) {
	// blocks is non-empty (skips the early return) but the template's block id is
	// not preserved, so the template's own body is retained (exists==false path).
	tpl := "<!-- centinela:keep:start:b -->\ntemplate body\n<!-- centinela:keep:end:b -->"
	out, kept := replaceKeepBlocks(tpl, map[string]string{"a": "preserved"})
	if kept != 0 {
		t.Fatalf("expected 0 preserved (id mismatch), got %d", kept)
	}
	if !strings.Contains(out, "template body") {
		t.Fatalf("expected template body retained, got %q", out)
	}
}

func TestReplaceKeepBlocksUnclosedTemplateBlock(t *testing.T) {
	// An unclosed block in the template: j reaches len(lines) so the i=j advance
	// is skipped, yet the preserved body is still substituted cleanly.
	tpl := "<!-- centinela:keep:start:a -->\nbody no end"
	out, kept := replaceKeepBlocks(tpl, map[string]string{"a": "kept-body"})
	if kept != 1 {
		t.Fatalf("expected preserved block, got %d", kept)
	}
	if !strings.Contains(out, "kept-body") {
		t.Fatalf("expected preserved body, got %q", out)
	}
}
