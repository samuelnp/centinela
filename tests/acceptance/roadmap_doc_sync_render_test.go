package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

func rdsRender(t *testing.T, js string) string {
	t.Helper()
	bin := buildCent(t)
	dir := rdsDir(t, js, rdsToml("warn"))
	return string(rdsGenerate(t, bin, dir))
}

// Scenario: Top-level intro blockquote round-trips from roadmap.json into ROADMAP.md
func TestRds_IntroBlockquote(t *testing.T) {
	out := rdsRender(t, `{"intro":"one\n\ntwo","phases":[]}`)
	mustHave(t, out, "# Roadmap\n\n")
	mustHave(t, out, "> one\n>\n> two")
}

// Scenario: Per-phase note renders as a blockquote preceding the feature list
func TestRds_PhaseNoteBlockquote(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","note":"a\n\nb","features":[{"name":"x"}]}]}`)
	mustHave(t, out, "## P\n\n> a\n>\n> b\n\n- **x**")
}

// Scenario: A phase with no note renders heading and features with no blockquote
func TestRds_PhaseNoNote(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x"}]}]}`)
	mustHave(t, out, "## P\n\n- **x**")
	if strings.Contains(out, ">") {
		t.Fatalf("no blockquote expected: %q", out)
	}
}

// Scenario: Feature with description and fixes renders both fields
func TestRds_FeatureDescriptionAndFixes(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x","description":"d","fixes":"f"}]}]}`)
	mustHave(t, out, "- **x** — d\n  *Fixes: f*")
}

// Scenario: Feature with description only renders no Fixes line
func TestRds_FeatureDescriptionOnly(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x","description":"d"}]}]}`)
	mustHave(t, out, "- **x** — d")
	if strings.Contains(out, "Fixes") {
		t.Fatalf("no Fixes line expected: %q", out)
	}
}

// Scenario: Feature with fixes only renders no em-dash clause on the bullet line
func TestRds_FeatureFixesOnly(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x","fixes":"f"}]}]}`)
	mustHave(t, out, "- **x**\n  *Fixes: f*")
}

// Scenario: Feature with no description and no fixes renders as a bare bullet with no dangling em-dash
func TestRds_FeatureBareBullet(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x"}]}]}`)
	mustHave(t, out, "- **x**")
	if strings.Contains(out, "- **x** —") {
		t.Fatalf("no dangling em-dash expected: %q", out)
	}
}

// Scenario: Feature with dependsOn renders dependency annotation in declared slice order
func TestRds_FeatureDependsOnOrder(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"feat-a"},{"name":"feat-b"},{"name":"x","description":"d","dependsOn":["feat-a","feat-b"]}]}]}`)
	mustHave(t, out, "- **x** — d (depends on feat-a, feat-b)")
}

// Scenario: Feature with dependsOn but no description attaches the annotation directly to the bullet
func TestRds_FeatureDependsOnNoDescription(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"feat-a"},{"name":"x","dependsOn":["feat-a"]}]}]}`)
	mustHave(t, out, "- **x** (depends on feat-a)")
}

// Scenario: Feature with empty dependsOn emits no dependency annotation
func TestRds_FeatureEmptyDependsOn(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x","dependsOn":[]}]}]}`)
	if strings.Contains(out, "depends on") {
		t.Fatalf("empty dependsOn must emit no clause: %q", out)
	}
}
