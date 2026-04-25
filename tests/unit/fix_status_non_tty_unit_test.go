package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusRunnerKeepsTTYFallbackLogic(t *testing.T) {
	path := filepath.Join("..", "..", "cmd", "centinela", "status_model.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read status model: %v", err)
	}
	content := string(data)
	checks := []string{
		"if !hasTTY(statusInput) || !hasTTY(statusOutput)",
		"fmt.Fprintln(statusOutput, renderStatusBody(wfs))",
		"tea.NewProgram(statusModel{workflows: wfs})",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("status runner missing %q", check)
		}
	}
}
