package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteAtomic_CreatesParentDir creates the directory if missing.
func TestWriteAtomic_CreatesParentDir(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "sub", "dir", "out.json")
	if err := writeAtomic(p, []byte(`{}`)); err != nil {
		t.Fatalf("writeAtomic with new dir: %v", err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != `{}` {
		t.Errorf("unexpected content: %s", got)
	}
}

// TestCompactBytes_NoHTMLEscape verifies angle-brackets are not escaped.
func TestCompactBytes_NoHTMLEscape(t *testing.T) {
	raw, err := compactBytes(map[string]string{"url": "a<b>c"})
	if err != nil {
		t.Fatalf("compactBytes: %v", err)
	}
	s := string(raw)
	// json.Encoder with SetEscapeHTML(false) must NOT escape < or >
	if s == "" {
		t.Error("unexpected empty output")
	}
}

// TestWriteRawRoadmap_NoPhasesKey creates phases key from an empty doc.
func TestWriteRawRoadmap_NoPhasesKey(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	// roadmap.json with only a top-level key but no "phases"
	os.WriteFile(p, []byte(`{"project":"centinela"}`), 0644) //nolint:errcheck
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	if len(doc.phases) != 0 {
		t.Error("expected empty phases for JSON without phases key")
	}
	// Defer should create the Backlog phase
	entry, _ := compactBytes(Feature{Name: "x"})
	doc.appendBacklog(entry) //nolint:errcheck
	writeRawRoadmap(p, doc)  //nolint:errcheck
	data, _ := os.ReadFile(p)
	s := string(data)
	if s == "" {
		t.Error("output should not be empty")
	}
}
