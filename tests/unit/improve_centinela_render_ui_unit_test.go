package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderUIBrandingPrimitivesExist(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "ui", "panel.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read panel file: %v", err)
	}
	content := string(data)
	checks := []string{"CENTINELA", "renderSystemPanel", "renderSystemLine", "toneWarn"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("panel missing %q", c)
		}
	}
}
