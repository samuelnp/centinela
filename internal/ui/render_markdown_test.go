package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
)

func mixedResults() []gates.Result {
	return []gates.Result{
		{Name: "G1: File Size", Status: gates.Fail, Message: "1 file over 100 lines",
			Details: []string{"internal/x.go (142 lines)"}},
		{Name: "import_graph", Status: gates.Pass, Message: "no forbidden edges"},
		{Name: "style", Status: gates.Warn, Message: "1 nit"},
		{Name: "i18n", Status: gates.Skip, Message: "no locale changes"},
	}
}

func TestRenderGatesMarkdown_MarkerAndFailDetails(t *testing.T) {
	out := RenderGatesMarkdown(mixedResults())
	if !strings.HasPrefix(out, MarkdownMarker+"\n") {
		t.Fatalf("marker must be the first line, got: %q", out[:40])
	}
	if !strings.Contains(out, "❌") {
		t.Fatalf("a failing result must render the fail icon: %q", out)
	}
	if !strings.Contains(out, "<details><summary>Failing details (G1: File Size)</summary>") {
		t.Fatalf("fail result must emit a <details> block: %q", out)
	}
	if !strings.Contains(out, "internal/x.go (142 lines)") {
		t.Fatalf("fail details must be rendered: %q", out)
	}
	// A passing gate must NOT get a <details> block.
	if strings.Contains(out, "Failing details (import_graph)") {
		t.Fatalf("passing gate must not produce <details>: %q", out)
	}
}

func TestRenderGatesMarkdown_AllPassHeader(t *testing.T) {
	out := RenderGatesMarkdown([]gates.Result{
		{Name: "G1: File Size", Status: gates.Pass, Message: "all under 100"},
	})
	if !strings.Contains(out, "✅") || !strings.Contains(out, "0 failed, 1 passed") {
		t.Fatalf("all-pass header wrong: %q", out)
	}
	if strings.Contains(out, "<details>") {
		t.Fatalf("all-pass output must have no <details>: %q", out)
	}
}

func TestRenderGatesMarkdown_WarnOnlyHeader(t *testing.T) {
	out := RenderGatesMarkdown([]gates.Result{
		{Name: "style", Status: gates.Warn, Message: "1 nit"},
	})
	if !strings.Contains(out, "⚠️") || !strings.Contains(out, "0 failed, 0 passed, 1 warned") {
		t.Fatalf("warn-only verdict must use the warn header icon: %q", out)
	}
}

func TestRenderGatesMarkdown_Deterministic(t *testing.T) {
	r := mixedResults()
	if RenderGatesMarkdown(r) != RenderGatesMarkdown(r) {
		t.Fatal("two renders over identical input must be byte-identical")
	}
}

func TestRenderGatesMarkdown_DetailsCap(t *testing.T) {
	big := make([]string, detailsCap+7)
	for i := range big {
		big[i] = "line"
	}
	out := RenderGatesMarkdown([]gates.Result{
		{Name: "G1: File Size", Status: gates.Fail, Message: "many", Details: big},
	})
	if !strings.Contains(out, "… 7 more") {
		t.Fatalf("oversized Details must be capped with an overflow note: %q", out)
	}
	if strings.Count(out, "- line") != detailsCap {
		t.Fatalf("exactly detailsCap detail lines must be shown, got %d", strings.Count(out, "- line"))
	}
}
