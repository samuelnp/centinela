package memory

import (
	"os"
	"testing"
	"time"
)

// SC-08/11: applyCaps truncates by count and byte limits.
func TestApplyCapsCountLimit(t *testing.T) {
	entries := make([]Entry, 20)
	for i := range entries {
		entries[i] = newEntry("g", "tests", TypeLesson, "- body", "s", nil, time.Now())
	}
	capped := applyCaps(entries, 5, 999999)
	if len(capped) != 5 {
		t.Fatalf("expected 5 entries after count cap, got %d", len(capped))
	}
}

// SC-11: byte cap halts before count cap.
func TestApplyCaps_ByteCapHalts(t *testing.T) {
	at := time.Now()
	e1 := newEntry("g", "tests", TypeLesson, "- body is exactly long enough to matter", "s", nil, at)
	e2 := newEntry("g", "tests", TypeVerdict, "- second entry body text", "s", nil, at)

	// Byte cap set to exactly fit e1 but not e1+e2.
	cap1Bytes := e1.sizeBytes()
	capped := applyCaps([]Entry{e1, e2}, 10, cap1Bytes)
	if len(capped) != 1 {
		t.Fatalf("expected 1 entry after byte cap, got %d", len(capped))
	}
}

// applyCaps with 0 entries returns empty.
func TestApplyCapsEmpty(t *testing.T) {
	if got := applyCaps(nil, 10, 9999); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

// FeatureTags extracts distinct tags from entries matching the feature.
func TestFeatureTags(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("beta", "tests", TypeLesson, "- body", "s", []string{"coverage", "lesson"}, time.Now())
	writeIfAbsent(e) //nolint:errcheck

	tags := FeatureTags("beta")
	if len(tags) == 0 {
		t.Fatal("expected tags for beta")
	}
	found := false
	for _, tag := range tags {
		if tag == "coverage" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'coverage' tag, got %v", tags)
	}
}

// FeatureTags excludes entries from other features.
func TestFeatureTagsExcludesOtherFeatures(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("other", "tests", TypeLesson, "- not for beta", "s", []string{"nope"}, time.Now())
	writeIfAbsent(e) //nolint:errcheck

	tags := FeatureTags("beta")
	if len(tags) != 0 {
		t.Fatalf("expected no tags for beta, got %v", tags)
	}
}

// toSet converts slice to membership map.
func TestToSet(t *testing.T) {
	m := toSet([]string{"a", "b", "a"})
	if !m["a"] || !m["b"] {
		t.Fatal("expected both keys present")
	}
	if len(m) != 2 {
		t.Fatalf("expected 2 entries (dedup), got %d", len(m))
	}
}
