package ui

import "testing"

func TestRenderStatusLineVariants(t *testing.T) {
	if got := RenderStatusLine(StatusLineView{}); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
	one := RenderStatusLine(StatusLineView{Primary: []string{"WF:alpha"}})
	if one != "WF:alpha" {
		t.Fatalf("unexpected single-line output: %q", one)
	}
	two := RenderStatusLine(StatusLineView{
		Primary:   []string{"WF:alpha", "STEP:code"},
		Secondary: []string{"NEXT:implement-code", "BLOCK:none"},
	})
	exp := "WF:alpha STEP:code\nNEXT:implement-code BLOCK:none"
	if two != exp {
		t.Fatalf("unexpected two-line output: %q", two)
	}
	onlySecond := RenderStatusLine(StatusLineView{Secondary: []string{"BLOCK:NO_WORKFLOW"}})
	if onlySecond != "BLOCK:NO_WORKFLOW" {
		t.Fatalf("unexpected secondary-only output: %q", onlySecond)
	}
}
