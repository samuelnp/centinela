package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoadmapAnalysisValidationAndGuardExist(t *testing.T) {
	analysisPath := filepath.Join("..", "..", "internal", "roadmap", "analysis.go")
	analysisData, err := os.ReadFile(analysisPath)
	if err != nil {
		t.Fatalf("expected file %s: %v", analysisPath, err)
	}
	analysisContent := string(analysisData)
	if !strings.Contains(analysisContent, "RoadmapAnalysisFile") {
		t.Fatalf("expected analysis constants in %s", analysisPath)
	}
	if !strings.Contains(analysisContent, "senior-product-manager") {
		t.Fatalf("expected senior PM role check in %s", analysisPath)
	}
	guardPath := filepath.Join("..", "..", "cmd", "centinela", "start_guard.go")
	guardData, err := os.ReadFile(guardPath)
	if err != nil {
		t.Fatalf("expected file %s: %v", guardPath, err)
	}
	if !strings.Contains(string(guardData), "requires roadmap senior PM analysis") {
		t.Fatalf("expected start guard analysis requirement in %s", guardPath)
	}
}
