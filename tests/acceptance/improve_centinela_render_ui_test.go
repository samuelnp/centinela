package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/improve-centinela-render-ui.feature
func TestHookOutputsUseCompactExplicitBrandedUI(t *testing.T) {
	paths := []string{
		filepath.Join("..", "..", "internal", "ui", "render.go"),
		filepath.Join("..", "..", "internal", "ui", "render_status.go"),
		filepath.Join("..", "..", "internal", "ui", "render_setup.go"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(data), "renderSystem") {
			t.Fatalf("expected branded renderer usage in %s", path)
		}
	}
}
