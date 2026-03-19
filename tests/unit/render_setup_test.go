package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
)

func TestRenderSetupNeeded_PointsToRoadmap(t *testing.T) {
	output := ui.RenderSetupNeeded()

	if strings.Contains(output, "centinela start") {
		t.Error("RenderSetupNeeded should not mention 'centinela start'")
	}
	if !strings.Contains(output, "roadmap") {
		t.Error("RenderSetupNeeded should mention 'roadmap'")
	}
}
