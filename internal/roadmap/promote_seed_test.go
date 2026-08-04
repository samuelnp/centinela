package roadmap

import (
	"os"
	"testing"
)

var seededPaths = []string{
	RoadmapAnalysisFile, RoadmapQualityFile,
	RoadmapAnalysisMarkdown, RoadmapQualityMarkdown,
}

// TestSeedArtifactsIfAbsentCreatesValidEmptyPair: the seeded artifacts must be
// structurally valid for the role checks and appendable by promote, or seeding
// would only move the failure one step later.
func TestSeedArtifactsIfAbsentCreatesValidEmptyPair(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("SeedArtifactsIfAbsent: %v", err)
	}
	for _, p := range seededPaths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("%s was not seeded: %v", p, err)
		}
	}
	// preflightArtifacts is what promote runs next; it must now pass.
	if err := preflightArtifacts(); err != nil {
		t.Fatalf("seeded artifacts must satisfy the promote preflight: %v", err)
	}
}

// TestSeedArtifactsIfAbsentNeverTouchesExistingFiles is the ❌ direction: a real
// evaluator pass must never be clobbered by a later promote.
func TestSeedArtifactsIfAbsentNeverTouchesExistingFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll(".workflow", 0755) //nolint:errcheck
	originals := map[string]string{}
	for _, p := range seededPaths {
		body := "PRE-EXISTING " + p
		os.WriteFile(p, []byte(body), 0644) //nolint:errcheck
		originals[p] = body
	}
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("SeedArtifactsIfAbsent: %v", err)
	}
	for p, want := range originals {
		got, _ := os.ReadFile(p)
		if string(got) != want {
			t.Fatalf("%s was rewritten: %q", p, got)
		}
	}
}

// TestSeedArtifactsIfAbsentIsIdempotentAndPartial: seeding twice changes
// nothing, and a half-present pair is completed rather than reset.
func TestSeedArtifactsIfAbsentIsIdempotentAndPartial(t *testing.T) {
	t.Chdir(t.TempDir())
	os.MkdirAll(".workflow", 0755)                              //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("mine"), 0644) //nolint:errcheck
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	first := map[string][]byte{}
	for _, p := range seededPaths {
		first[p], _ = os.ReadFile(p)
	}
	if err := SeedArtifactsIfAbsent(); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	for _, p := range seededPaths {
		again, _ := os.ReadFile(p)
		if string(again) != string(first[p]) {
			t.Fatalf("%s is not idempotent", p)
		}
	}
	if string(first[RoadmapAnalysisMarkdown]) != "mine" {
		t.Fatal("the pre-existing half of the pair must survive")
	}
}

// TestSeedArtifactsIfAbsentSurfacesWriteFailures: an unwritable .workflow path
// must produce an error, not a silent half-seed that promote then trips over.
func TestSeedArtifactsIfAbsentSurfacesWriteFailures(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(".workflow", []byte("not a directory"), 0644) //nolint:errcheck
	if err := SeedArtifactsIfAbsent(); err == nil {
		t.Fatal("a write failure must surface")
	}
}
