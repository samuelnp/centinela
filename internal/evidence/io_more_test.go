package evidence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestWriteAtomicNilEvidence(t *testing.T) {
	chdirToTemp(t)
	if err := WriteAtomic("alpha", orchestration.RoleBigThinker, nil); err == nil {
		t.Fatal("expected nil-evidence error")
	}
}

func TestWriteBytesAtomicCreatesAndOverwrites(t *testing.T) {
	d := chdirToTemp(t)
	target := filepath.Join(d, ".workflow", "raw.txt")
	if err := writeBytesAtomic(target, []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := writeBytesAtomic(target, []byte("second")); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "second" {
		t.Fatalf("expected overwrite, got %q", got)
	}
}

func TestWriteBytesAtomicPropagatesOpenError(t *testing.T) {
	chdirToTemp(t)
	err := writeBytesAtomic(".workflow/nested/missing/foo", []byte("x"))
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestReadParsesAndPreservesExtras(t *testing.T) {
	chdirToTemp(t)
	raw := []byte(`{"feature":"alpha","step":"plan","role":"big-thinker","status":"done","generatedAt":"2026-05-12T00:00:00Z","inputs":["a"],"outputs":["b"],"edgeCases":[],"handoffTo":"feature-specialist","x":1}`)
	if err := os.WriteFile(pathFor("alpha", orchestration.RoleBigThinker), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := Read("alpha", orchestration.RoleBigThinker)
	if err != nil {
		t.Fatal(err)
	}
	if string(r.Extra["x"]) != "1" {
		t.Fatalf("extras not preserved: %v", r.Extra)
	}
	if _, err := json.Marshal(r); err != nil {
		t.Fatal(err)
	}
}

func TestReadCorruptedJSONErrors(t *testing.T) {
	chdirToTemp(t)
	if err := os.WriteFile(pathFor("alpha", orchestration.RoleBigThinker), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read("alpha", orchestration.RoleBigThinker); err == nil {
		t.Fatal("expected parse error")
	}
}
