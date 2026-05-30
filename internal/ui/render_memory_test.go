package ui

import (
	"strings"
	"testing"
)

// TestRenderMemoryBlockEmpty — no facts produces empty string.
func TestRenderMemoryBlockEmpty(t *testing.T) {
	got := RenderMemoryBlock(nil)
	if got != "" {
		t.Fatalf("expected empty string for no facts, got %q", got)
	}
}

// TestRenderMemoryBlockEmptySlice — empty slice also produces empty string.
func TestRenderMemoryBlockEmptySlice(t *testing.T) {
	got := RenderMemoryBlock([]string{})
	if got != "" {
		t.Fatalf("expected empty string for empty slice, got %q", got)
	}
}

// TestRenderMemoryBlockSingleFact — header and fact line are both present.
func TestRenderMemoryBlockSingleFact(t *testing.T) {
	got := RenderMemoryBlock([]string{"alpha [lesson]: timeout edge case"})
	if got == "" {
		t.Fatal("expected non-empty block for a single fact")
	}
	if !strings.Contains(got, "MEMORY") {
		t.Fatalf("expected MEMORY header, got %q", got)
	}
	if !strings.Contains(got, "alpha [lesson]: timeout edge case") {
		t.Fatalf("expected fact in block, got %q", got)
	}
}

// TestRenderMemoryBlockMultipleFacts — all facts appear.
func TestRenderMemoryBlockMultipleFacts(t *testing.T) {
	facts := []string{
		"alpha [lesson]: lesson one",
		"beta [verdict]: verdict one",
	}
	got := RenderMemoryBlock(facts)
	for _, f := range facts {
		if !strings.Contains(got, f) {
			t.Fatalf("expected fact %q in block, got %q", f, got)
		}
	}
}
