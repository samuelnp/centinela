package docgen

import (
	"strings"
	"testing"
)

func TestRenderKBIndex_WrittenAndPlaceholder(t *testing.T) {
	d := &Data{
		Title:   "Centinela",
		Specs:   []string{"specs/alpha.feature", "specs/beta.feature"},
		States:  []FeatureState{{Feature: "alpha", Status: "done"}, {Feature: "beta", Status: "in-progress"}},
		KB: []KBPage{{
			Feature: "alpha", Summary: "Plain summary.", Status: "done",
			WhatItDoes: "x", WhenToUse: "y", HowItBehaves: "z",
		}},
	}
	html := RenderKBIndex(d)
	for _, want := range []string{"Knowledge Base", "alpha", "Plain summary.", "beta", "Guide not yet written", "Status: done", "Status: in-progress"} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in KB index", want)
		}
	}
	if !strings.Contains(html, `href="alpha.html"`) {
		t.Fatal("expected link to alpha.html")
	}
	if strings.Contains(html, `href="beta.html"`) {
		t.Fatal("beta has no md; should not link to a page")
	}
}

func TestRenderKBFeature_AllSectionsAndOptionalExamples(t *testing.T) {
	p := KBPage{
		Feature: "alpha", Summary: "S.", Status: "done",
		WhatItDoes: "does it", WhenToUse: "when need", HowItBehaves: "- one\n- two",
	}
	html := RenderKBFeature(p, "Centinela")
	for _, want := range []string{"alpha", "What it does", "does it", "When you'd use it", "when need", "How it behaves", "<li>one</li>", "<li>two</li>"} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in KB feature page", want)
		}
	}
	if strings.Contains(html, `<section id="examples"`) {
		t.Fatal("Examples section should be omitted when empty")
	}
	p.Examples = "Run centinela docs."
	with := RenderKBFeature(p, "Centinela")
	if !strings.Contains(with, "Run centinela docs.") || !strings.Contains(with, `<section id="examples"`) {
		t.Fatal("Examples not rendered when present")
	}
}

func TestRenderKBFeature_FallbackSummary(t *testing.T) {
	p := KBPage{Feature: "alpha", WhatItDoes: "x", WhenToUse: "y", HowItBehaves: "z"}
	html := RenderKBFeature(p, "Centinela")
	if !strings.Contains(html, "End-user guide for this feature.") {
		t.Fatal("expected fallback summary in hero")
	}
}
