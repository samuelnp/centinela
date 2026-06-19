package synthesize

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

func inv(lang string, pkgs []string, manifests ...analyze.Manifest) analyze.Inventory {
	return analyze.Inventory{
		SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: lang,
		Packages: pkgs, Manifests: manifests,
	}
}

func TestInfer_Archetypes(t *testing.T) {
	cases := []struct {
		name string
		in   analyze.Inventory
		want Archetype
		conf string
	}{
		{"n-tier", inv("Go", []string{"internal/handler", "internal/service", "internal/repository"}), NTier, High},
		{"rails", inv("Ruby", []string{"app/models", "app/controllers", "app/views"},
			analyze.Manifest{Kind: "gem", Path: "Gemfile", Deps: []string{"rails"}}), RailsNative, High},
		{"ecs", inv("GDScript", []string{"src/systems", "src/components", "src/entities"}), ECS, High},
		{"hexagonal", inv("Go", []string{"domain", "application", "ports", "infrastructure"}), Hexagonal, High},
		{"modular", inv("Go", []string{"modules/billing/public", "modules/billing/internal"}), Modular, Medium},
		{"empty", inv("", nil), Custom, Low},
	}
	for _, c := range cases {
		got := NewInferer().Infer(c.in)
		if got.Best != c.want {
			t.Errorf("%s: best=%s want %s (scores=%+v)", c.name, got.Best, c.want, got.Scores)
		}
		if got.Confidence != c.conf {
			t.Errorf("%s: confidence=%s want %s", c.name, got.Confidence, c.conf)
		}
	}
}

func TestInfer_AmbiguousTie(t *testing.T) {
	// "service" (n-tier +2) and "domain" (hexagonal +2) tie at margin 0.
	got := NewInferer().Infer(inv("Go", []string{"service", "domain"}))
	if !got.Ambiguous || got.Confidence != Low {
		t.Fatalf("expected ambiguous low, got ambiguous=%v conf=%s scores=%+v", got.Ambiguous, got.Confidence, got.Scores)
	}
}

func TestInfer_DeterministicOrderAndReasons(t *testing.T) {
	in := inv("Go", []string{"internal/handler", "internal/service", "internal/repository"})
	a := NewInferer().Infer(in)
	b := NewInferer().Infer(in)
	if a.Scores[0].Archetype != b.Scores[0].Archetype || len(a.Reasons()) == 0 {
		t.Fatalf("inference must be deterministic with reasons: %+v", a)
	}
	if NewInferer().Infer(inv("", nil)).Reasons() != nil {
		t.Fatal("custom fallback should have no reasons")
	}
}
