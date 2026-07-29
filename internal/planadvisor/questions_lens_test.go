package planadvisor

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// D9: every question is tagged with a LENS name, never a role slug.
func TestSelectQuestions_LensTagsAreStrategyOrSpec(t *testing.T) {
	qs := selectQuestions(bundle{}, 99, "always")
	if len(qs) == 0 {
		t.Fatal("always mode must produce questions")
	}
	seen := map[string]bool{}
	for _, q := range qs {
		if q.Lens != "strategy" && q.Lens != "spec" {
			t.Errorf("question tagged with a non-lens value %q: %s", q.Lens, q.Text)
		}
		seen[q.Lens] = true
	}
	if !seen["strategy"] || !seen["spec"] {
		t.Fatalf("both lenses must be represented, saw %v", seen)
	}
}

func TestSelectQuestions_NoLegacyRoleSlugsRemain(t *testing.T) {
	for _, q := range selectQuestions(bundle{}, 99, "always") {
		for _, retired := range []string{"big-thinker", "feature-specialist"} {
			if q.Lens == retired {
				t.Fatalf("retired role slug %q still used as a lens tag", retired)
			}
		}
	}
}

// The advisor header states one agent with two lenses, not two agents.
func TestDirective_HeaderNamesOnePlannerTwoLenses(t *testing.T) {
	out := Directive("nonexistent-feature", &config.Config{})
	if !strings.Contains(out, "One planner agent, two lenses: strategy first, then spec.") {
		t.Fatalf("advisor header missing the one-agent phrasing: %s", out)
	}
	for _, retired := range []string{"big-thinker", "feature-specialist"} {
		if strings.Contains(out, retired) {
			t.Fatalf("advisor output must not name %q: %s", retired, out)
		}
	}
}

// Rendering shape is unchanged: `- [<lens>] <text>`.
func TestDirective_RendersBracketedLensTags(t *testing.T) {
	out := Directive("nonexistent-feature", &config.Config{})
	if !strings.Contains(out, "- [strategy] ") && !strings.Contains(out, "- [spec] ") {
		t.Fatalf("expected bracketed lens tags in the rendered output: %s", out)
	}
}

// A nil config must not panic and must still render the one-agent header.
func TestDirective_NilConfigUsesDefaults(t *testing.T) {
	out := Directive("nonexistent-feature", nil)
	if !strings.Contains(out, "One planner agent, two lenses") {
		t.Fatalf("nil config must fall back to defaults: %s", out)
	}
}

// Fully covered planning docs produce no questions — the header still holds.
func TestDirective_NoQuestionsStillNamesOneAgent(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll("docs/features", 0755) //nolint:errcheck
	body := "## Problem\nx\n## Scope\nx\n## Acceptance Criteria\nx\n## Edge Cases\nx\n" +
		"## Constraints\nx\n## Risks\nx\n"
	os.WriteFile("docs/features/f.md", []byte(body), 0644) //nolint:errcheck
	out := Directive("f", &config.Config{})
	if !strings.Contains(out, "One planner agent, two lenses") {
		t.Fatalf("header must always render: %s", out)
	}
	if strings.Contains(out, "[big-thinker]") || strings.Contains(out, "[feature-specialist]") {
		t.Fatalf("no retired lens tags may appear: %s", out)
	}
}
