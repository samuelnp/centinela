package unit_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func indexHTMLPath(t *testing.T) string {
	t.Helper()
	_, f, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(f), "..", "..")
	return filepath.Join(root, "web", "index.html")
}

// web/index.html exists and is non-trivial (> 1 KB).
func TestLandingPageUnit_FileExists(t *testing.T) {
	p := indexHTMLPath(t)
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("web/index.html missing: %v", err)
	}
	if info.Size() < 1024 {
		t.Errorf("web/index.html suspiciously small (%d bytes)", info.Size())
	}
}

// Single <h1> element — heading hierarchy is correct.
func TestLandingPageUnit_SingleH1(t *testing.T) {
	data, err := os.ReadFile(indexHTMLPath(t))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	html := string(data)
	count := strings.Count(html, "<h1")
	if count != 1 {
		t.Errorf("expected exactly 1 <h1>, got %d", count)
	}
}

// Viewport meta tag is present for mobile-first rendering.
func TestLandingPageUnit_ViewportMeta(t *testing.T) {
	data, _ := os.ReadFile(indexHTMLPath(t))
	html := string(data)
	if !strings.Contains(html, "width=device-width") {
		t.Error("viewport meta 'width=device-width' not found")
	}
}

// theme-color meta is present.
func TestLandingPageUnit_ThemeColor(t *testing.T) {
	data, _ := os.ReadFile(indexHTMLPath(t))
	html := string(data)
	if !strings.Contains(html, `name="theme-color"`) {
		t.Error("theme-color meta tag not found")
	}
}
