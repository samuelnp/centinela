package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersonaPrimitivesExist(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "ui", "persona.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persona file: %v", err)
	}
	content := string(data)
	checks := []string{"personaLabel", "🛡️👁️"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("persona file missing %q", c)
		}
	}
}
