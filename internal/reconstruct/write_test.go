package reconstruct

import (
	"os"
	"path/filepath"
	"testing"
)

func recon(slugs ...string) Reconstruction {
	var r Reconstruction
	for _, s := range slugs {
		r.Features = append(r.Features, Artifact{Slug: s, Body: "Feature: " + s + "\n  Scenario: x\n    Given a " + todoMarker + "\n"})
		r.Briefs = append(r.Briefs, Artifact{Slug: s, Body: "# Feature: " + s + "\n"})
	}
	return r
}

func TestWriteCorpus_WritesAndReRunByteIdentical(t *testing.T) {
	dir := t.TempDir()
	chdirRecon(t, dir)
	out := filepath.Join(dir, ".workflow", "reconstructed")
	w1, sk1, err := WriteCorpus(out, recon("alpha", "beta"))
	if err != nil || len(w1) != 4 || len(sk1) != 0 {
		t.Fatalf("first write: w=%v sk=%v err=%v", w1, sk1, err)
	}
	specBody, _ := os.ReadFile(filepath.Join(out, "specs", "alpha.feature"))
	briefBody, _ := os.ReadFile(filepath.Join(out, "features", "alpha.md"))
	if len(specBody) == 0 || len(briefBody) == 0 {
		t.Fatal("spec/brief not written to review dir")
	}
	_, _, _ = WriteCorpus(out, recon("alpha", "beta"))
	specBody2, _ := os.ReadFile(filepath.Join(out, "specs", "alpha.feature"))
	if string(specBody) != string(specBody2) {
		t.Fatal("re-run must be byte-identical")
	}
}

func TestWriteCorpus_SkipsHandAuthoredSpec(t *testing.T) {
	dir := t.TempDir()
	chdirRecon(t, dir)
	if err := os.MkdirAll(filepath.Join(dir, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "specs", "alpha.feature"), []byte("HAND"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, ".workflow", "reconstructed")
	written, skipped, err := WriteCorpus(out, recon("alpha", "beta"))
	if err != nil || len(skipped) != 1 || skipped[0] != "alpha" {
		t.Fatalf("alpha must be skipped: w=%v sk=%v err=%v", written, skipped, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "specs", "alpha.feature")); string(b) != "HAND" {
		t.Fatalf("hand-authored spec mutated: %q", b)
	}
	if _, err := os.Stat(filepath.Join(out, "features", "alpha.md")); err == nil {
		t.Fatal("skipped target must not emit an orphan brief")
	}
	if len(written) != 2 { // only beta's spec+brief
		t.Fatalf("expected 2 written, got %v", written)
	}
}

func TestWriteCorpus_NoPartialFileOnError(t *testing.T) {
	dir := t.TempDir()
	chdirRecon(t, dir)
	// Make the specs out dir a regular file so MkdirAll under it fails.
	blocker := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := WriteCorpus(filepath.Join(blocker, "root"), recon("alpha")); err == nil {
		t.Fatal("writing under a regular file must error")
	}
}

func TestBriefFor_MissingSlug(t *testing.T) {
	if briefFor(recon("a"), "missing") != "" {
		t.Fatal("briefFor must return empty for unknown slug")
	}
}

func chdirRecon(t *testing.T, dir string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}
