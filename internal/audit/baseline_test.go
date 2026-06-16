package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func sampleBaseline() Baseline {
	return Baseline{Scheme: fingerprintScheme, Version: 1, Gates: []GateEntry{
		{Gate: "G1: File Size", Fingerprints: Compute("G1: File Size", []string{"b.go (5 lines)", "a.go (9 lines)"})},
	}}
}

// TestSaveLoadRoundTrip writes then reads a baseline (through a missing nested
// parent dir, exercising MkdirAll) and gets the same content back.
func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dir", "baseline.json")
	if err := Save(path, sampleBaseline()); err != nil {
		t.Fatal(err)
	}
	got, exists, err := Load(path)
	if err != nil || !exists {
		t.Fatalf("load: exists=%v err=%v", exists, err)
	}
	if got.Scheme != fingerprintScheme || got.Version != 1 || len(got.Gates) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if len(got.Gates[0].Fingerprints) != 2 {
		t.Fatalf("want 2 fingerprints, got %d", len(got.Gates[0].Fingerprints))
	}
}

// TestLoadMissing returns exists=false (no error) for an absent file.
func TestLoadMissing(t *testing.T) {
	_, exists, err := Load(filepath.Join(t.TempDir(), "none.json"))
	if err != nil || exists {
		t.Fatalf("missing file: exists=%v err=%v", exists, err)
	}
}

// TestSaveDeterministic writes the same baseline twice and gets byte-identical
// output (AC-7), regardless of input fingerprint order.
func TestSaveDeterministic(t *testing.T) {
	p1 := filepath.Join(t.TempDir(), "a.json")
	p2 := filepath.Join(t.TempDir(), "b.json")
	if Save(p1, sampleBaseline()) != nil || Save(p2, sampleBaseline()) != nil {
		t.Fatal("save failed")
	}
	d1, _ := os.ReadFile(p1)
	d2, _ := os.ReadFile(p2)
	if string(d1) != string(d2) {
		t.Fatal("re-record not byte-identical")
	}
	if d1[len(d1)-1] != '\n' {
		t.Fatal("missing trailing newline")
	}
}

// TestSaveSorts confirms gates by name and fingerprints by hash on disk.
func TestSaveSorts(t *testing.T) {
	b := Baseline{Scheme: fingerprintScheme, Version: 1, Gates: []GateEntry{
		{Gate: "import_graph", Fingerprints: Compute("import_graph", []string{"z (x)"})},
		{Gate: "G1: File Size", Fingerprints: Compute("G1: File Size", []string{"a.go (1 lines)", "m.go (2 lines)", "q.go (3 lines)"})},
	}}
	path := filepath.Join(t.TempDir(), "x.json")
	if err := Save(path, b); err != nil {
		t.Fatal(err)
	}
	got, _, _ := Load(path)
	if got.Gates[0].Gate != "G1: File Size" {
		t.Fatalf("gates not sorted: %q first", got.Gates[0].Gate)
	}
	fps := got.Gates[0].Fingerprints
	for i := 1; i < len(fps); i++ {
		if fps[i-1].Hash > fps[i].Hash {
			t.Fatal("fingerprints not sorted by hash")
		}
	}
}

// TestSchemeStale flags a baseline recorded under a different scheme.
func TestSchemeStale(t *testing.T) {
	if (Baseline{Scheme: fingerprintScheme}).SchemeStale() {
		t.Fatal("current scheme should not be stale")
	}
	if !(Baseline{Scheme: "v0"}).SchemeStale() {
		t.Fatal("mismatched scheme should be stale")
	}
	if (Baseline{}).SchemeStale() {
		t.Fatal("empty scheme should not be stale")
	}
}
