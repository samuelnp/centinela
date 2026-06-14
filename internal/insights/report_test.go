package insights

import "testing"

// rankTop must sort by count desc, then key asc, and truncate to n.
func TestRankTopSortsCountDescKeyAsc(t *testing.T) {
	m := map[string]int{"b": 1, "a": 1, "c": 3, "d": 2}
	got := rankTop(m, 10)
	want := []Count{{"c", 3}, {"d", 2}, {"a", 1}, {"b", 1}}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pos %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// n smaller than the bucket count truncates; n larger returns all; n<=0 empty.
func TestRankTopTruncation(t *testing.T) {
	m := map[string]int{"a": 5, "b": 4, "c": 3}
	if got := rankTop(m, 2); len(got) != 2 || got[0].Key != "a" || got[1].Key != "b" {
		t.Fatalf("top 2 = %+v", got)
	}
	if got := rankTop(m, 10); len(got) != 3 {
		t.Fatalf("top 10 of 3 = %d entries, want 3", len(got))
	}
	if got := rankTop(m, 0); len(got) != 0 {
		t.Fatalf("top 0 = %d entries, want 0", len(got))
	}
	if got := rankTop(m, -1); len(got) != 0 {
		t.Fatalf("top -1 = %d entries, want 0", len(got))
	}
}

// An empty map yields an empty (non-nil-typed) slice without panic.
func TestRankTopEmptyMap(t *testing.T) {
	if got := rankTop(map[string]int{}, 5); len(got) != 0 {
		t.Fatalf("empty map = %+v, want []", got)
	}
}
