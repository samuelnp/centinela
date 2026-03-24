package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectStageParserAndBootstrapHelpersExist(t *testing.T) {
	stagePath := filepath.Join("..", "..", "internal", "projectstage", "stage.go")
	stageData, err := os.ReadFile(stagePath)
	if err != nil {
		t.Fatalf("expected file %s: %v", stagePath, err)
	}
	if !strings.Contains(string(stageData), "greenfield") {
		t.Fatalf("expected greenfield support in %s", stagePath)
	}
	bootstrapPath := filepath.Join("..", "..", "internal", "roadmap", "bootstrap.go")
	bootstrapData, err := os.ReadFile(bootstrapPath)
	if err != nil {
		t.Fatalf("expected file %s: %v", bootstrapPath, err)
	}
	if !strings.Contains(strings.ToLower(string(bootstrapData)), "phase 0") {
		t.Fatalf("expected bootstrap phase detection in %s", bootstrapPath)
	}
}
