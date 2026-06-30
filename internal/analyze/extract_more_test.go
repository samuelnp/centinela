package analyze

import "testing"

func TestFirstQuoted_Variants(t *testing.T) {
	cases := map[string]string{
		`gem "rails"`:    "rails",
		"gem 'rspec'":    "rspec",
		"no quotes here": "",
		`unterminated "`: "", // opening quote with no closing match
	}
	for in, want := range cases {
		if got := firstQuoted(in); got != want {
			t.Fatalf("firstQuoted(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNpmFramework_NoneAndSecondaryMatch(t *testing.T) {
	// No known framework dependency -> empty string.
	if got := npmFramework(map[string]string{"lodash": "4"}); got != "" {
		t.Fatalf("expected no framework, got %q", got)
	}
	// A non-first framework in the priority list still resolves.
	if got := npmFramework(map[string]string{"express": "4"}); got != "Express" {
		t.Fatalf("expected Express, got %q", got)
	}
}

func TestSortedDepNames_EmptyReturnsNil(t *testing.T) {
	if got := sortedDepNames(map[string]string{}, nil); got != nil {
		t.Fatalf("expected nil for no deps, got %#v", got)
	}
}
