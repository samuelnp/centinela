package reconstruct

import (
	"strings"
	"testing"
)

func TestSelect_ExclusionWinsOverPromotion(t *testing.T) {
	in := inv("Go", []string{
		"internal/service_test", // test-only
		"vendor/foo/handler",    // generated/vendored
		"internal/config",       // config leaf
		"pkg/mocks/service",     // mocks
		"internal/service",      // survives
	})
	got := Select(in)
	if len(got) != 1 || got[0].Pkg != "internal/service" {
		t.Fatalf("exclusion must drop test/vendor/config/mocks, kept: %v", slugs(got))
	}
}

func TestSelect_SlugCollisionDisambiguated(t *testing.T) {
	// "a/service" and "b/service" both slugify around "service" but differ via
	// path; identical leaf "service" twice forces a numeric suffix.
	in := inv("Go", []string{"service", "service/"})
	got := Select(in)
	if len(got) != 2 {
		t.Fatalf("both packages must select, got %v", slugs(got))
	}
	if got[0].Slug == got[1].Slug {
		t.Fatalf("colliding slugs must be disambiguated: %v", slugs(got))
	}
	if !strings.HasPrefix(got[1].Slug, "service") {
		t.Fatalf("disambiguated slug should keep base: %v", slugs(got))
	}
}

func TestSelect_BoundedToMaxTargets(t *testing.T) {
	var pkgs []string
	for i := 0; i < maxTargets+25; i++ {
		pkgs = append(pkgs, "internal/service"+itoa(i))
	}
	got := Select(inv("Go", pkgs))
	if len(got) != maxTargets {
		t.Fatalf("target count must be bounded to %d, got %d", maxTargets, len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Slug >= got[i].Slug {
			t.Fatalf("bounded set must stay slug-sorted at %d", i)
		}
	}
}
