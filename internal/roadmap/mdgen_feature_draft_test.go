package roadmap

import (
	"strings"
	"testing"
)

// TestRenderFeature_DraftMarker appends a deterministic trailing " *(draft)*".
func TestRenderFeature_DraftMarker(t *testing.T) {
	got := strings.Join(renderFeature(Feature{Name: "x", Draft: true}), "\n")
	if got != "- **x** *(draft)*" {
		t.Fatalf("bare draft bullet: %q", got)
	}
	full := strings.Join(renderFeature(Feature{
		Name: "x", Description: "d", DependsOn: []string{"a"}, Draft: true,
	}), "\n")
	if full != "- **x** — d (depends on a) *(draft)*" {
		t.Fatalf("draft marker must trail the whole bullet: %q", full)
	}
	// A non-draft never gets the marker.
	if strings.Contains(strings.Join(renderFeature(Feature{Name: "x"}), "\n"), "draft") {
		t.Fatal("non-draft must not carry the marker")
	}
}
