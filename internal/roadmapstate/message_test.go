package roadmapstate

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMessageIsConventional(t *testing.T) {
	if got := Message("defer", "flaky-thing"); got != "chore(roadmap): defer flaky-thing" {
		t.Fatalf("Message = %q", got)
	}
	if got := Message("phase add", "Phase 14"); got != "chore(roadmap): phase add Phase 14" {
		t.Fatalf("multi-word verb: %q", got)
	}
}

func TestMessageWithoutSubject(t *testing.T) {
	for _, subject := range []string{"", "   ", "\n\t"} {
		if got := Message("generate", subject); got != "chore(roadmap): generate" {
			t.Fatalf("Message(%q) = %q", subject, got)
		}
	}
}

func TestMessageCollapsesNewlines(t *testing.T) {
	got := Message("phase rename", "old\nname   with\tgaps")
	if got != "chore(roadmap): phase rename old name with gaps" {
		t.Fatalf("Message = %q", got)
	}
	if strings.ContainsAny(got, "\n\t") {
		t.Fatal("a commit subject must stay one line")
	}
}

func TestMessageTruncatesLongSubjects(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := Message("edit", long)
	subject := strings.TrimPrefix(got, "chore(roadmap): edit ")
	if n := utf8.RuneCountInString(subject); n != SubjectLimit {
		t.Fatalf("subject is %d runes, want %d", n, SubjectLimit)
	}
	if !strings.HasSuffix(subject, "…") {
		t.Fatalf("a truncated subject must be marked: %q", subject)
	}
}

// Truncation slices runes, never bytes: a multi-byte subject must stay valid.
func TestMessageTruncatesOnRuneBoundaries(t *testing.T) {
	got := Message("edit", strings.Repeat("é", 200))
	if !utf8.ValidString(got) {
		t.Fatalf("truncation produced invalid UTF-8: %q", got)
	}
}

func TestMessageKeepsSubjectAtTheLimit(t *testing.T) {
	exact := strings.Repeat("y", SubjectLimit)
	if got := Message("move", exact); got != "chore(roadmap): move "+exact {
		t.Fatalf("a subject exactly at the limit must not be cut: %q", got)
	}
}
