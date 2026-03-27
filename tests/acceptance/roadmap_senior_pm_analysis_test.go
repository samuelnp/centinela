package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-senior-pm-analysis.feature
func TestRoadmapSeniorPMAnalysisHookAndCommandWiring(t *testing.T) {
	setupPath := filepath.Join("..", "..", "cmd", "centinela", "hook_setup.go")
	setupData, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read hook setup file: %v", err)
	}
	setupContent := string(setupData)
	if !strings.Contains(setupContent, "roadmap analysis required") {
		t.Fatal("roadmap analysis setup directive missing")
	}
	validatePath := filepath.Join("..", "..", "cmd", "centinela", "roadmap_validate.go")
	validateData, err := os.ReadFile(validatePath)
	if err != nil {
		t.Fatalf("read roadmap validate file: %v", err)
	}
	if !strings.Contains(string(validateData), "roadmapCmd.AddCommand(roadmapValidateCmd)") {
		t.Fatal("roadmap validate subcommand wiring missing")
	}
}
