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

func TestRenderSetupNeeded_AsksExactSetupQuestions(t *testing.T) {
	output := ui.RenderSetupNeeded()
	checks := []string{
		"Project name - what should we call it?",
		"Elevator pitch - one sentence: what does it do and for whom?",
		"Tech stack - language, framework, styling, persistence, test tools?",
		"Architecture archetype - hexagonal, rails-native, n-tier, ecs, modular, or custom?",
		"Locales - which languages does the UI support? (default: English only)",
		"Folder layout - preferred structure, or should I propose one based on the archetype?",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("RenderSetupNeeded missing %q", check)
		}
	}
}
