package roadmap

import (
	"strings"
	"testing"
)

func renderFeatureStr(f Feature) string {
	return strings.Join(renderFeature(f), "\n")
}

func TestRenderFeatureBothFields(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", Description: "d", Fixes: "fx"})
	if got != "- **x** — d\n  *Fixes: fx*" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderFeatureDescriptionOnly(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", Description: "d"})
	if got != "- **x** — d" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "Fixes") {
		t.Fatalf("no Fixes line expected: %q", got)
	}
}

func TestRenderFeatureFixesOnly(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", Fixes: "fx"})
	if got != "- **x**\n  *Fixes: fx*" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, " — ") {
		t.Fatalf("no em-dash clause expected: %q", got)
	}
}

func TestRenderFeatureBareBullet(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x"})
	if got != "- **x**" {
		t.Fatalf("bare bullet expected, got %q", got)
	}
}

func TestRenderFeatureDependsOnWithDescription(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", Description: "d", DependsOn: []string{"a", "b"}})
	if got != "- **x** — d (depends on a, b)" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderFeatureDependsOnNoDescription(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", DependsOn: []string{"a"}})
	if got != "- **x** (depends on a)" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderFeatureEmptyDependsOn(t *testing.T) {
	got := renderFeatureStr(Feature{Name: "x", DependsOn: []string{}})
	if strings.Contains(got, "depends on") {
		t.Fatalf("empty dependsOn must emit no clause: %q", got)
	}
}

func TestRenderBacklogFeatureFull(t *testing.T) {
	got := strings.Join(renderBacklogFeature(Feature{Name: "x", Summary: "s",
		DeferredAt: "2026-01-01", Source: &Source{Feature: "feat", Role: "qa"}}), "\n")
	if got != "- **x** — s *(deferred 2026-01-01 · feat/qa)*" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderBacklogFeatureNoSource(t *testing.T) {
	got := strings.Join(renderBacklogFeature(Feature{Name: "x", Summary: "s"}), "\n")
	if got != "- **x** — s" {
		t.Fatalf("got %q", got)
	}
	for _, bad := range []string{"()", "· /", "*("} {
		if strings.Contains(got, bad) {
			t.Fatalf("must not contain %q: %q", bad, got)
		}
	}
}

func TestBacklogProvenanceHalves(t *testing.T) {
	if got := backlogProvenance(&Source{Feature: "f"}); got != "f" {
		t.Fatalf("feature-only: %q", got)
	}
	if got := backlogProvenance(&Source{Role: "r"}); got != "r" {
		t.Fatalf("role-only: %q", got)
	}
	if got := backlogProvenance(nil); got != "" {
		t.Fatalf("nil source: %q", got)
	}
}
