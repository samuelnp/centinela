package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteArtifact_SortedKeys writes non-features keys in sorted order.
func TestWriteArtifact_SortedKeys(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "art.json")
	top := map[string]json.RawMessage{
		"role":     json.RawMessage(`"pm"`),
		"zzz":      json.RawMessage(`"last"`),
		"aaa":      json.RawMessage(`"first"`),
		"features": json.RawMessage(`[]`),
	}
	if err := writeArtifact(p, top); err != nil {
		t.Fatalf("writeArtifact: %v", err)
	}
	data, _ := os.ReadFile(p)
	s := string(data)
	zzz := strings.Index(s, "zzz")
	aaa := strings.Index(s, "aaa")
	if aaa == -1 || zzz == -1 || aaa > zzz {
		t.Errorf("sorted key order violated: aaa=%d zzz=%d in:\n%s", aaa, zzz, s)
	}
}

// TestWriteArtifact_ByteStableAcrossRuns two identical calls produce equal output.
func TestWriteArtifact_ByteStableAcrossRuns(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "art.json")
	top := map[string]json.RawMessage{
		"role":     json.RawMessage(`"pm"`),
		"features": json.RawMessage(`[{"name":"x"}]`),
	}
	writeArtifact(p, top) //nolint:errcheck
	run1, _ := os.ReadFile(p)
	writeArtifact(p, top) //nolint:errcheck
	run2, _ := os.ReadFile(p)
	if string(run1) != string(run2) {
		t.Errorf("writeArtifact must be byte-stable:\nrun1=%s\nrun2=%s", run1, run2)
	}
}

// TestWriteFeatureArray_MultipleEntries emits one line per entry.
func TestWriteFeatureArray_MultipleEntries(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "art.json")
	top := map[string]json.RawMessage{
		"features": json.RawMessage(`[{"name":"a"},{"name":"b"}]`),
	}
	writeArtifact(p, top) //nolint:errcheck
	data, _ := os.ReadFile(p)
	s := string(data)
	lineA := strings.Count(s, `"a"`)
	lineB := strings.Count(s, `"b"`)
	if lineA != 1 || lineB != 1 {
		t.Errorf("each feature must appear exactly once: %s", s)
	}
}

// TestAppendLine_CreatesFile creates a file when absent.
func TestAppendLine_CreatesFile(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "notes.md")
	if err := appendLine(p, "- first bullet"); err != nil {
		t.Fatalf("appendLine: %v", err)
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "first bullet") {
		t.Errorf("line not written: %s", data)
	}
}

// TestAppendLine_AppendsToExisting preserves prior content.
func TestAppendLine_AppendsToExisting(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "notes.md")
	os.WriteFile(p, []byte("existing content\n"), 0644) //nolint:errcheck
	appendLine(p, "- new bullet")                       //nolint:errcheck
	data, _ := os.ReadFile(p)
	s := string(data)
	if !strings.Contains(s, "existing content") {
		t.Error("prior content must be preserved")
	}
	if !strings.Contains(s, "new bullet") {
		t.Error("new bullet must be appended")
	}
}
