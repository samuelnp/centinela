package analyze

import "testing"

func TestDetectLanguages_MapsAndSumsExtensions(t *testing.T) {
	stats, primary := detectLanguages(map[string]int{
		".go": 3, ".js": 1, ".jsx": 1, ".unknown": 99,
	})
	if primary != "Go" {
		t.Fatalf("primary: got %q want Go", primary)
	}
	got := map[string]int{}
	for _, s := range stats {
		got[s.Name] = s.FileCount
	}
	if got["Go"] != 3 || got["JavaScript"] != 2 {
		t.Fatalf("counts: %#v", got)
	}
	if _, ok := got["unknown"]; ok {
		t.Fatal("unknown extension must not be counted as a language")
	}
}

func TestDetectLanguages_SortsCountDescNameAsc(t *testing.T) {
	// Equal counts (Go vs Ruby) must break alphabetically (deterministic).
	stats, primary := detectLanguages(map[string]int{".go": 2, ".rb": 2, ".py": 5})
	if primary != "Python" {
		t.Fatalf("primary: got %q want Python", primary)
	}
	if stats[0].Name != "Python" || stats[1].Name != "Go" || stats[2].Name != "Ruby" {
		t.Fatalf("order: %#v", stats)
	}
}

func TestDetectLanguages_EmptyInput(t *testing.T) {
	stats, primary := detectLanguages(map[string]int{})
	if primary != "" || len(stats) != 0 {
		t.Fatalf("empty input must yield empty stats and primary \"\": %q %#v", primary, stats)
	}
}
