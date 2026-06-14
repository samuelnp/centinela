package roadmap

import (
	"bytes"
	"strings"
	"testing"
)

func mustRender(t *testing.T, r *Roadmap) string {
	t.Helper()
	return string(RenderMarkdown(r))
}

// Golden: a full roadmap with intro, phase note, feature with both fields, and
// a Backlog phase renders to the exact canonical byte sequence.
func TestRenderMarkdownGolden(t *testing.T) {
	r := &Roadmap{
		Intro: "Principle line.\n\nStatus line.",
		Phases: []Phase{
			{Name: "✅ Phase 0: Bootstrap", Note: "Para one.\n\nPara two.",
				Features: []Feature{{Name: "setup", Description: "Wire it up.", Fixes: "broken hook"}}},
			{Name: "Backlog", Features: []Feature{
				{Name: "f-defer", Summary: "deferred bit", DeferredAt: "2026-01-01",
					Source: &Source{Feature: "feat", Role: "qa"}}}},
		},
	}
	want := "# Roadmap\n\n> Principle line.\n>\n> Status line.\n\n" +
		"## ✅ Phase 0: Bootstrap\n\n> Para one.\n>\n> Para two.\n\n" +
		"- **setup** — Wire it up.\n  *Fixes: broken hook*\n\n" +
		"## Backlog\n\n- **f-defer** — deferred bit *(deferred 2026-01-01 · feat/qa)*\n"
	if got := mustRender(t, r); got != want {
		t.Fatalf("golden mismatch\n got:%q\nwant:%q", got, want)
	}
}

// Determinism: rendering the same roadmap twice yields byte-identical output.
func TestRenderMarkdownDeterministic(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "P", Features: []Feature{{Name: "a"}, {Name: "b"}}}}}
	a := RenderMarkdown(r)
	b := RenderMarkdown(r)
	if !bytes.Equal(a, b) {
		t.Fatalf("non-deterministic: %q vs %q", a, b)
	}
}

// EOF contract: exactly one trailing newline, no trailing whitespace, LF only.
func TestRenderMarkdownTrailingNewline(t *testing.T) {
	out := mustRender(t, &Roadmap{Phases: []Phase{{Name: "P", Features: []Feature{{Name: "a"}}}}})
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Fatalf("want exactly one trailing newline, got %q", out)
	}
	if strings.Contains(out, "\r") {
		t.Fatalf("output must be LF-only, found CR: %q", out)
	}
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if l != strings.TrimRight(l, " \t") {
			t.Fatalf("trailing whitespace on line %q", l)
		}
	}
}

// Intro multi-line with a blank inner line emits "> " lines and a bare ">".
func TestRenderMarkdownIntroBlockquote(t *testing.T) {
	out := mustRender(t, &Roadmap{Intro: "one\n\ntwo", Phases: nil})
	want := "# Roadmap\n\n> one\n>\n> two\n"
	if out != want {
		t.Fatalf("intro mismatch\n got:%q\nwant:%q", out, want)
	}
}

// Non-ASCII prose passes through byte-for-byte.
func TestRenderMarkdownNonASCII(t *testing.T) {
	out := mustRender(t, &Roadmap{Phases: []Phase{{Name: "P",
		Features: []Feature{{Name: "x", Description: "café — “quote” ✅"}}}}})
	if !strings.Contains(out, "café — “quote” ✅") {
		t.Fatalf("non-ASCII not preserved: %q", out)
	}
}
