package roadmap

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// render must surface an error when an untouched phase holds invalid JSON.
func TestRender_PhaseIndentError(t *testing.T) {
	doc := newRawDoc(`{bad}`)
	if _, err := doc.render(); err == nil {
		t.Fatal("expected render error for invalid phase bytes")
	}
}

// writeRawRoadmap must surface a render error before touching the filesystem.
func TestWriteRawRoadmap_RenderErrorPropagates(t *testing.T) {
	doc := newRawDoc(`{bad}`)
	p := filepath.Join(t.TempDir(), "roadmap.json")
	if err := writeRawRoadmap(p, doc); err == nil {
		t.Fatal("expected writeRawRoadmap to fail on render error")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("no file should be written when render fails")
	}
}

// renderDirtyPhase must surface an error when "features" is not an array.
func TestRenderDirtyPhase_FeaturesNotArray(t *testing.T) {
	if _, err := renderDirtyPhase(json.RawMessage(`{"name":"x","features":42}`)); err == nil {
		t.Fatal("expected renderDirtyPhase error for non-array features")
	}
}

// writePhaseKey prefixes a comma when it is not the first key emitted.
func TestWritePhaseKey_NotFirst(t *testing.T) {
	var buf bytes.Buffer
	first := false
	writePhaseKey(&buf, "name", json.RawMessage(`"x"`), &first)
	if !strings.HasPrefix(buf.String(), ",") {
		t.Fatalf("expected leading comma for non-first key, got %q", buf.String())
	}
}

// writeAtomic must surface a CreateTemp error when the parent dir is read-only.
func TestWriteAtomic_CreateTempError(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("read-only dir enforcement unavailable")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) }) //nolint:errcheck
	// MkdirAll on the existing dir succeeds; CreateTemp inside it must fail.
	if err := writeAtomic(filepath.Join(dir, "out.json"), []byte("{}")); err == nil {
		t.Fatal("expected CreateTemp error in a read-only directory")
	}
}
