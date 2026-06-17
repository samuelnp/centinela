package memory

import (
	"testing"
	"time"
)

// TestMarshalUnmarshalRoundTrip — full field round-trip through markdown+frontmatter.
func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	at := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	in := Entry{
		ID:             "abc123",
		Feature:        "alpha",
		Step:           "tests",
		Type:           TypeLesson,
		Title:          "timeout lesson",
		Tags:           []string{"edge-cases", "lesson"},
		SourceArtifact: ".workflow/alpha-edge-cases.md",
		CreatedAt:      at,
		Body:           "- timeout on retry",
	}
	data := marshal(in)
	out, ok := unmarshal(data)
	if !ok {
		t.Fatalf("unmarshal failed; data:\n%s", data)
	}
	if out.ID != in.ID {
		t.Fatalf("ID mismatch: %q != %q", out.ID, in.ID)
	}
	if out.Feature != in.Feature {
		t.Fatalf("Feature mismatch")
	}
	if out.Type != in.Type {
		t.Fatalf("Type mismatch")
	}
	if out.Body != in.Body {
		t.Fatalf("Body mismatch: %q != %q", out.Body, in.Body)
	}
	if !out.CreatedAt.Equal(in.CreatedAt) {
		t.Fatalf("CreatedAt mismatch: %v != %v", out.CreatedAt, in.CreatedAt)
	}
	if len(out.Tags) != 2 || out.Tags[0] != "edge-cases" {
		t.Fatalf("Tags mismatch: %v", out.Tags)
	}
}

// TestUnmarshalMissingFrontmatterDelimiter returns false.
func TestUnmarshalMissingFrontmatterDelimiter(t *testing.T) {
	_, ok := unmarshal([]byte("no frontmatter here"))
	if ok {
		t.Fatal("expected unmarshal to fail for missing delimiter")
	}
}

// TestUnmarshalNoClosingFence returns false.
func TestUnmarshalNoClosingFence(t *testing.T) {
	_, ok := unmarshal([]byte("---\nid: abc\nfeature: f\n"))
	if ok {
		t.Fatal("expected unmarshal to fail for unclosed frontmatter")
	}
}

// TestUnmarshalMissingID returns false.
func TestUnmarshalMissingID(t *testing.T) {
	data := []byte("---\nfeature: f\ntype: lesson\n---\nbody\n")
	_, ok := unmarshal(data)
	if ok {
		t.Fatal("expected unmarshal to fail when id is missing")
	}
}

// TestSplitTagsEmpty returns empty slice for blank input.
func TestSplitTagsEmpty(t *testing.T) {
	tags := splitTags("")
	if len(tags) != 0 {
		t.Fatalf("expected empty tags, got %v", tags)
	}
}

// TestSplitTagsTrimsWhitespace verifies trimming.
func TestSplitTagsTrimsWhitespace(t *testing.T) {
	tags := splitTags("  a , b  , c  ")
	if len(tags) != 3 || tags[0] != "a" || tags[2] != "c" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

// TestAssignFieldBadTimestampGraceful — bad createdAt results in zero time, not panic.
func TestAssignFieldBadTimestampGraceful(t *testing.T) {
	var e Entry
	assignField(&e, "createdAt", "not-a-date")
	if !e.CreatedAt.IsZero() {
		t.Fatal("expected zero time for bad timestamp")
	}
}
