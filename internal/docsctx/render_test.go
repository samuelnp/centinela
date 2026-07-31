package docsctx

import (
	"strings"
	"testing"
)

func renderedContext(changelogPresent bool) Context {
	ctx := Context{
		Feature: "f",
		Brief:   Section{Source: "docs/features/f.md", Body: "brief body\n", Present: true},
		Plan:    Section{Source: "docs/plans/f.md", Body: "plan body\n", Present: true},
		Spec:    Section{Source: "specs/f.feature", Body: "Feature: f\n", Present: true},
	}
	ctx.Changelog = Section{Source: ".workflow/f-changelog.md"}
	if changelogPresent {
		ctx.Changelog = Section{Source: ".workflow/f-changelog.md", Body: "- feat: f\n", Present: true}
	}
	return ctx
}

// The output is plain markdown: title, one section per input, each with a
// `> source:` line and the verbatim body.
func TestRenderSectionsAndSources(t *testing.T) {
	out := Render(renderedContext(true))
	if !strings.HasPrefix(out, "# Docs Context: f\n") {
		t.Fatalf("missing title: %s", out)
	}
	for _, want := range []string{
		"## Feature brief", "> source: docs/features/f.md", "brief body",
		"## Plan", "> source: docs/plans/f.md", "plan body",
		"## Spec scenarios", "> source: specs/f.feature", "Feature: f",
		"## Changelog draft", "> source: .workflow/f-changelog.md", "- feat: f",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "no changelog draft yet") {
		t.Fatalf("present changelog must not render the hint:\n%s", out)
	}
}

// An absent changelog draft renders the actionable hint and keeps the section.
func TestRenderAbsentChangelogHint(t *testing.T) {
	out := Render(renderedContext(false))
	if !strings.Contains(out, "## Changelog draft") {
		t.Fatalf("changelog section must always render:\n%s", out)
	}
	if !strings.Contains(out, "no changelog draft yet — run: centinela artifact new f changelog") {
		t.Fatalf("absent changelog must render the artifact-new hint:\n%s", out)
	}
}

// Bodies embed with trailing newlines trimmed to exactly one.
func TestRenderTrimsTrailingNewlines(t *testing.T) {
	ctx := renderedContext(true)
	ctx.Brief.Body = "brief body\n\n\n"
	out := Render(ctx)
	if !strings.Contains(out, "brief body\n\n## Plan") {
		t.Fatalf("trailing newlines must collapse before the next section:\n%s", out)
	}
}

// Determinism: rendering the same context twice is byte-identical.
func TestRenderDeterministic(t *testing.T) {
	first := Render(renderedContext(true))
	second := Render(renderedContext(true))
	if first != second {
		t.Fatal("Render must be deterministic")
	}
}
