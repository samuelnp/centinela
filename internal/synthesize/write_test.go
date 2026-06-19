package synthesize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDraft_FreshWritesTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "PROJECT.md")
	written, clobbered, err := WriteDraft(target, "body")
	if err != nil || clobbered || written != target {
		t.Fatalf("fresh write: %q clobbered=%v err=%v", written, clobbered, err)
	}
	if b, _ := os.ReadFile(target); string(b) != "body" {
		t.Fatalf("content not written: %q", b)
	}
}

func TestWriteDraft_NeverClobbersExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "PROJECT.md")
	if err := os.WriteFile(target, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}
	written, clobbered, err := WriteDraft(target, "new draft")
	if err != nil || !clobbered {
		t.Fatalf("existing target must not be clobbered: clobbered=%v err=%v", clobbered, err)
	}
	if written != filepath.Join(dir, "PROJECT.draft.md") {
		t.Fatalf("draft path wrong: %q", written)
	}
	if b, _ := os.ReadFile(target); string(b) != "ORIGINAL" {
		t.Fatalf("original PROJECT.md mutated: %q", b)
	}
	if b, _ := os.ReadFile(written); string(b) != "new draft" {
		t.Fatalf("draft content wrong: %q", b)
	}
}

func TestWriteDraft_UnwritableDirErrors(t *testing.T) {
	if _, _, err := WriteDraft(filepath.Join(t.TempDir(), "nope", "x", "PROJECT.md"), "x"); err == nil {
		// MkdirAll creates parents, so use a path under a file to force failure.
	}
	f := filepath.Join(t.TempDir(), "afile")
	os.WriteFile(f, []byte("x"), 0o644)
	if _, _, err := WriteDraft(filepath.Join(f, "PROJECT.md"), "x"); err == nil {
		t.Fatal("writing under a regular file must error")
	}
}

func TestDraftPathFor(t *testing.T) {
	if got := draftPathFor("PROJECT.md"); got != "PROJECT.draft.md" {
		t.Fatalf("draftPathFor: %q", got)
	}
	if got := draftPathFor(filepath.Join("a", "PROJECT.md")); got != filepath.Join("a", "PROJECT.draft.md") {
		t.Fatalf("draftPathFor nested: %q", got)
	}
}
