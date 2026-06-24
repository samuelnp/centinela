package audit

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAdoptByteIdenticalToRecordSave: the file Adopt writes is byte-identical to
// an independent Record+Save on the same unchanged repo — adopt adds semantics,
// not different data.
func TestAdoptByteIdenticalToRecordSave(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	o, err := Adopt(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	adopted, _ := os.ReadFile(o.Path)
	ref := filepath.Join(t.TempDir(), "ref.json")
	if err := Save(ref, Record(cfg)); err != nil {
		t.Fatal(err)
	}
	refBytes, _ := os.ReadFile(ref)
	if string(adopted) != string(refBytes) {
		t.Fatal("adopt output differs from Record+Save reference")
	}
}

// TestAdoptCleanRepoZeroFindings: a repo with no violations yields a written,
// zero-finding baseline (nothing to ratchet).
func TestAdoptCleanRepoZeroFindings(t *testing.T) {
	cfg := tempRepo(t, "fail", nil)
	o, err := Adopt(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if o.Skipped || o.Baseline.Total() != 0 {
		t.Fatalf("clean repo: skipped=%v total=%d", o.Skipped, o.Baseline.Total())
	}
	if _, err := os.Stat(o.Path); err != nil {
		t.Fatalf("zero-finding baseline not written: %v", err)
	}
}

// TestAdoptLoadErrorPropagates: a baseline path that is a directory makes Load
// fail (not a missing-file), and Adopt surfaces the error without writing.
func TestAdoptLoadErrorPropagates(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	path := cfg.Gates.AuditBaseline.BaselinePath
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Adopt(cfg, false); err == nil {
		t.Fatal("expected Adopt to propagate the Load error")
	}
}

// TestBaselineTotal counts fingerprints across gates.
func TestBaselineTotal(t *testing.T) {
	b := Baseline{Gates: []GateEntry{
		{Gate: "G1: File Size", Fingerprints: Compute("G1: File Size", []string{"a.go (5 lines)", "b.go (9 lines)"})},
		{Gate: "import_graph", Fingerprints: Compute("import_graph", []string{"x → y (z)"})},
	}}
	if got := b.Total(); got != 3 {
		t.Fatalf("Total = %d, want 3", got)
	}
	if (Baseline{}).Total() != 0 {
		t.Fatal("empty baseline Total should be 0")
	}
}
