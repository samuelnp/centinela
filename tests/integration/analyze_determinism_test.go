package integration_test

import (
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

// snapshotSources returns sha256 of every non-output file under root, so a
// read-only run can be proven not to mutate any source file (AC-6).
func snapshotSources(t *testing.T, root, outRel string) map[string][32]byte {
	t.Helper()
	sums := map[string][32]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if rel == outRel {
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		sums[rel] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return sums
}

func TestAnalyzeIsByteIdenticalAndReadOnly(t *testing.T) {
	root := writeAnalyzeFixture(t)
	before := snapshotSources(t, root, analyze.DefaultOutPath)

	first := analyzeIn(t, root)
	second := analyzeIn(t, root)
	if string(first) != string(second) {
		t.Fatalf("re-run must be byte-identical (AC-3/4)")
	}

	after := snapshotSources(t, root, analyze.DefaultOutPath)
	if len(before) != len(after) {
		t.Fatalf("source file set changed: %d -> %d", len(before), len(after))
	}
	for rel, sum := range before {
		if after[rel] != sum {
			t.Fatalf("read-only violated: %s was mutated (AC-6)", rel)
		}
	}
}
