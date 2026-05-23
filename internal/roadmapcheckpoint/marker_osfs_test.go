package roadmapcheckpoint

import (
	"os"
	"testing"
	"time"
)

func TestMarkerRoundTripAndOSFS(t *testing.T) {
	chdirTmp(t)

	if m, err := ReadMarker(MarkerPath); m != nil || err != nil {
		t.Fatalf("missing marker -> (nil,nil), got (%+v,%v)", m, err)
	}

	if err := WriteMarker(MarkerPath, t0); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}
	m, err := ReadMarker(MarkerPath)
	if err != nil || m == nil || m.Choice != "iterate" {
		t.Fatalf("ReadMarker -> (%+v, %v)", m, err)
	}
	if _, err := time.Parse(time.RFC3339, m.At); err != nil {
		t.Fatalf("marker At not RFC3339: %q", m.At)
	}

	// Malformed marker on disk -> ReadMarker returns an error.
	if err := os.WriteFile(MarkerPath, []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMarker(MarkerPath); err == nil {
		t.Fatal("malformed marker should error from ReadMarker")
	}

	// Unparseable at -> ReadMarker returns marker AND error.
	if err := os.WriteFile(MarkerPath, marker("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if mm, err := ReadMarker(MarkerPath); err == nil || mm == nil {
		t.Fatalf("unparseable at should return (marker, err), got (%+v, %v)", mm, err)
	}

	// NewOSFS exercises Stat/ReadFile/Exists against real disk.
	fs := NewOSFS()
	if !fs.Exists(MarkerPath) {
		t.Fatal("OSFS.Exists should see the marker")
	}
	if _, ok := fs.Stat(MarkerPath); !ok {
		t.Fatal("OSFS.Stat should see the marker")
	}
	if _, ok := fs.ReadFile(MarkerPath); !ok {
		t.Fatal("OSFS.ReadFile should read the marker")
	}
	if fs.Exists("nonexistent-path-xyz") {
		t.Fatal("OSFS.Exists should be false for missing path")
	}
	if _, ok := fs.Stat("nonexistent-path-xyz"); ok {
		t.Fatal("OSFS.Stat should be false for missing path")
	}
	if _, ok := fs.ReadFile("nonexistent-path-xyz"); ok {
		t.Fatal("OSFS.ReadFile should be false for missing path")
	}
}
