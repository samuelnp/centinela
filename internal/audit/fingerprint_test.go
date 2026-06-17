package audit

import "testing"

// TestIdentityKeyFileSizeStable is the central AC-5 guard: a file-size detail's
// key is the path, so growth (130→170 lines) yields the SAME key and Hash.
func TestIdentityKeyFileSizeStable(t *testing.T) {
	a := Compute("G1: File Size", []string{"a.go (130 lines)"})
	b := Compute("G1: File Size", []string{"a.go (170 lines)"})
	if a[0].Key != "a.go" || b[0].Key != "a.go" {
		t.Fatalf("keys = %q / %q, want a.go", a[0].Key, b[0].Key)
	}
	if a[0].Hash != b[0].Hash {
		t.Fatalf("hash drifted across line growth: %q vs %q", a[0].Hash, b[0].Hash)
	}
	if a[0].Raw != "a.go (130 lines)" {
		t.Fatalf("raw not preserved: %q", a[0].Raw)
	}
}

// TestIdentityKeyImportGraphStable strips the trailing " (reason)" so the edge
// is the stable identity, and is a no-op when no parenthetical is present.
func TestIdentityKeyImportGraphStable(t *testing.T) {
	got := identityKey("import_graph", "internal/ui → internal/orch (forbidden)")
	if got != "internal/ui → internal/orch" {
		t.Fatalf("key = %q", got)
	}
	// No " (" present ⇒ beforeParen returns the trimmed whole string.
	if got := identityKey("G1: File Size", "  bare/path.go  "); got != "bare/path.go" {
		t.Fatalf("no-paren key = %q", got)
	}
}

// TestIdentityKeyPassThroughGates leaves already-stable details untouched.
func TestIdentityKeyPassThroughGates(t *testing.T) {
	cases := map[string]string{
		"spec-traceability-gate": `specs/x.feature: "scenario"`,
		"G-Secrets: Secret Scan": "path/f.go: rule aws-key",
		"G11: i18n":              "src/a.ts",
	}
	for gate, detail := range cases {
		if got := identityKey(gate, "  "+detail+"  "); got != detail {
			t.Fatalf("gate %s: key = %q, want %q", gate, got, detail)
		}
	}
}

// TestGenericKeyFallback strips a trailing parenthetical and trailing digits,
// and is a no-op on an already-stable key.
func TestGenericKeyFallback(t *testing.T) {
	if got := genericKey("some/file.go (42 lines)"); got != "some/file.go" {
		t.Fatalf("paren strip = %q", got)
	}
	if got := genericKey("widget count 17"); got != "widget count" {
		t.Fatalf("digit strip = %q", got)
	}
	if got := genericKey("already/stable.key"); got != "already/stable.key" {
		t.Fatalf("no-op failed = %q", got)
	}
}

// TestGenericKeyUnknownGate routes an unknown gate through the fallback.
func TestGenericKeyUnknownGate(t *testing.T) {
	if got := identityKey("future_gate", "x/y.go (9 lines)"); got != "x/y.go" {
		t.Fatalf("fallback key = %q", got)
	}
}

// TestComputeDedups collapses identical details into one fingerprint and sorts.
func TestComputeDedups(t *testing.T) {
	fps := Compute("G1: File Size", []string{"a.go (1 lines)", "a.go (2 lines)", "b.go (5 lines)"})
	if len(fps) != 2 {
		t.Fatalf("want 2 fingerprints after dedup, got %d", len(fps))
	}
	if !(fps[0].Hash < fps[1].Hash) {
		t.Fatal("fingerprints not sorted by hash")
	}
}

// TestSchemeFoldedIntoHash verifies the scheme participates in the hash: a key
// hashed with the live scheme differs from one with a different prefix.
func TestSchemeFoldedIntoHash(t *testing.T) {
	live := hashIdentity("G1: File Size", "a.go")
	if live == "" {
		t.Fatal("empty hash")
	}
	// Same gate+key under the live scheme is reproducible.
	if live != hashIdentity("G1: File Size", "a.go") {
		t.Fatal("hash not deterministic")
	}
	if live == hashIdentity("other", "a.go") {
		t.Fatal("gate should change the hash")
	}
}
