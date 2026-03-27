package docgen

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func ValidateInputs() error {
	if _, err := os.Stat("PROJECT.md"); err != nil {
		return fmt.Errorf("missing PROJECT.md")
	}
	if _, err := os.Stat("ROADMAP.md"); err != nil {
		return fmt.Errorf("missing ROADMAP.md")
	}
	r, err := roadmap.Load()
	if err != nil || r == nil || len(r.Phases) == 0 {
		return fmt.Errorf("missing .workflow/roadmap.json")
	}
	md := exists(roadmap.RoadmapAnalysisMarkdown)
	js := exists(roadmap.RoadmapAnalysisFile)
	if md || js {
		if err := roadmap.ValidateAnalysis(r); err != nil {
			return fmt.Errorf("invalid roadmap analysis: %w", err)
		}
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
