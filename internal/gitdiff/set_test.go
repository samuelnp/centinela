package gitdiff

import "testing"

func TestSetContainsLenAndNilSafety(t *testing.T) {
	s := NewSet([]string{"internal/gates/file_size.go", "cmd/centinela/validate.go", ""})
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
	if !s.Contains("internal/gates/file_size.go") {
		t.Fatalf("expected to contain forward-slash path")
	}
	if s.Contains("missing.go") {
		t.Fatalf("did not expect membership for missing.go")
	}

	var nilSet *Set
	if nilSet.Contains("x") || nilSet.Len() != 0 || nilSet.HasPrefix("x") {
		t.Fatalf("nil receiver methods must be safe and return zero values")
	}
}

func TestSetHasPrefix(t *testing.T) {
	s := NewSet([]string{"src/i18n/messages/en.json", "src/app/main.ts"})
	if !s.HasPrefix("src/i18n/messages/") {
		t.Fatalf("expected prefix match for locales dir")
	}
	if !s.HasPrefix("src/i18n/messages") {
		t.Fatalf("partial prefix should also match")
	}
	if s.HasPrefix("src/locales/") {
		t.Fatalf("unrelated prefix should not match")
	}
	if s.HasPrefix("src/i18n/messages/en.json/extra/path") {
		t.Fatalf("prefix longer than any key should not match")
	}
}
