package autostart

import (
	"strings"
	"testing"
)

func TestExtractPrompt_Branches(t *testing.T) {
	if got := ExtractPrompt([]byte("  plain text  ")); got != "plain text" {
		t.Fatalf("plain: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"prompt":"hello"}`)); got != "hello" {
		t.Fatalf("prompt key: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"input":{"text":"nested"}}`)); got != "nested" {
		t.Fatalf("nested input.text: %q", got)
	}
	if got := ExtractPrompt([]byte(`{not json`)); got != "{not json" {
		t.Fatalf("invalid json should return raw text: %q", got)
	}
	if got := ExtractPrompt([]byte(`{"other":"x"}`)); got != `{"other":"x"}` {
		t.Fatalf("no known key should return raw text: %q", got)
	}
}

func TestShouldStart_Branches(t *testing.T) {
	if ShouldStart("") {
		t.Fatal("empty must be false")
	}
	if ShouldStart("Shall I advance to code?") {
		t.Fatal("advance prompt must be false")
	}
	if !ShouldStart("I want to add worktrees") {
		t.Fatal("'i want to add' must be true")
	}
	if ShouldStart("just a normal sentence") {
		t.Fatal("non-intent must be false")
	}
}

func TestDeriveFeature_Fallbacks(t *testing.T) {
	// All stop words / too short -> timestamp fallback.
	got := DeriveFeature("i want to add a new feature")
	if !strings.HasPrefix(got, "feature-") {
		t.Fatalf("all-stopword prompt should fall back to feature-<ts>, got %q", got)
	}
	// Leading-digit derived name gets a feature- prefix.
	got = DeriveFeature("please add 2fa authentication support")
	if !strings.HasPrefix(got, "feature-") {
		t.Fatalf("leading-digit name should be prefixed, got %q", got)
	}
	// Normal multi-word prompt -> kebab.
	got = DeriveFeature("please add parallel worktrees isolation")
	if got != "parallel-worktrees-isolation" {
		t.Fatalf("kebab derivation: %q", got)
	}
}
