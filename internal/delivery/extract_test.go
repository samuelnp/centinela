package delivery

import "testing"

const extSrc = "intro\n\n## Problem\n\nthe problem body\nmore\n\n## Who / Why\n\nwhy text\n"

func TestExtractSectionFound(t *testing.T) {
	if got := ExtractSection(extSrc, "Problem"); got != "the problem body\nmore" {
		t.Fatalf("found: %q", got)
	}
}

func TestExtractSectionCaseInsensitive(t *testing.T) {
	if got := ExtractSection(extSrc, "  pRoBlEm  "); got != "the problem body\nmore" {
		t.Fatalf("case-insensitive: %q", got)
	}
}

func TestExtractSectionLast(t *testing.T) {
	if got := ExtractSection(extSrc, "Who / Why"); got != "why text" {
		t.Fatalf("last section: %q", got)
	}
}

func TestExtractSectionMissing(t *testing.T) {
	if got := ExtractSection(extSrc, "Nope"); got != "" {
		t.Fatalf("missing should be empty: %q", got)
	}
}

func TestExtractSectionEmptyInputs(t *testing.T) {
	if ExtractSection("", "Problem") != "" {
		t.Fatal("empty src")
	}
	if ExtractSection(extSrc, "  ") != "" {
		t.Fatal("empty heading")
	}
}

func TestHeadingTextRejectsDeep(t *testing.T) {
	if _, ok := headingText("### Deep"); ok {
		t.Fatal("### should not be a top-level heading")
	}
	if _, ok := headingText("plain"); ok {
		t.Fatal("non-heading")
	}
}

func TestFirstParagraph(t *testing.T) {
	if got := FirstParagraph("  one\nline\n\nsecond\n"); got != "one\nline" {
		t.Fatalf("multi: %q", got)
	}
	if got := FirstParagraph("just one"); got != "just one" {
		t.Fatalf("single: %q", got)
	}
	if FirstParagraph("   \n  ") != "" {
		t.Fatal("blank should be empty")
	}
}
