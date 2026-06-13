package roadmap

import (
	"encoding/json"
	"testing"
)

// TestSortedKeys returns keys in ascending lexicographic order.
func TestSortedKeys(t *testing.T) {
	m := map[string]json.RawMessage{
		"role":     json.RawMessage(`"pm"`),
		"zebra":    json.RawMessage(`1`),
		"alpha":    json.RawMessage(`2`),
		"features": json.RawMessage(`[]`),
	}
	got := sortedKeys(m)
	want := []string{"alpha", "features", "role", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("sortedKeys len %d, want %d", len(got), len(want))
	}
	for i, k := range got {
		if k != want[i] {
			t.Errorf("sortedKeys[%d] = %q, want %q", i, k, want[i])
		}
	}
}

// TestSortedKeys_Empty handles an empty map.
func TestSortedKeys_Empty(t *testing.T) {
	if got := sortedKeys(map[string]json.RawMessage{}); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}
