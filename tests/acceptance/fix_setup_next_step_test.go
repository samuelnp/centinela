package acceptance_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
)

// Acceptance: spec/fix-setup-next-step.feature
// Scenario: After writing PROJECT.md, Claude guides user to roadmap
func TestSetupNextStep_NoMentionOfCentinelaStart(t *testing.T) {
	output := ui.RenderSetupNeeded()
	if strings.Contains(output, "centinela start") {
		t.Error("closing instruction must not mention 'centinela start <feature>'")
	}
}

func TestSetupNextStep_MentionsRoadmap(t *testing.T) {
	output := ui.RenderSetupNeeded()
	if !strings.Contains(output, "roadmap") {
		t.Error("closing instruction must mention 'roadmap'")
	}
}
