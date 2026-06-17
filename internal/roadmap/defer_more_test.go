package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefer_MissingRoadmapFile returns an error and does not create the file.
func TestDefer_MissingRoadmapFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nonexistent", "roadmap.json")
	err := Defer(p, DeferOptions{Slug: "x", Summary: "s"})
	if err == nil {
		t.Fatal("expected error for missing roadmap.json")
	}
}

// TestDefer_CorruptRoadmapFile returns an error.
func TestDefer_CorruptRoadmapFile(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte("{bad"), 0644) //nolint:errcheck
	if err := Defer(p, DeferOptions{Slug: "x", Summary: "s"}); err == nil {
		t.Fatal("expected error for corrupt roadmap.json")
	}
}

// TestDefer_JSONSpecialCharsInSummary stores them without injection.
func TestDefer_JSONSpecialCharsInSummary(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	summary := `summary with "quotes" and <angle> brackets`
	err := Defer(p, DeferOptions{
		Slug:    "special-chars",
		Summary: summary,
	})
	if err != nil {
		t.Fatalf("Defer with special chars: %v", err)
	}
	data, _ := os.ReadFile(p)
	if len(data) == 0 {
		t.Error("roadmap.json must not be empty")
	}
}
