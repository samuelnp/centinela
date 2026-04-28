package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
)

func TestRenderSetupNeeded_OutputIsNonEmpty(t *testing.T) {
	output := ui.RenderSetupNeeded()
	if strings.TrimSpace(output) == "" {
		t.Error("RenderSetupNeeded returned empty output")
	}
}

func TestRenderSetupNeeded_ContainsProjectMDReady(t *testing.T) {
	output := ui.RenderSetupNeeded()
	if !strings.Contains(output, "PROJECT.md is ready") {
		t.Error("expected 'PROJECT.md is ready' in output")
	}
}

func TestRenderSetupNeeded_RequiresExactQuestions(t *testing.T) {
	output := ui.RenderSetupNeeded()
	if !strings.Contains(output, "Ask these exact questions; do not combine or omit them") {
		t.Error("expected exact setup question directive")
	}
	if !strings.Contains(output, "Project name") || !strings.Contains(output, "Folder layout") {
		t.Error("expected setup checklist endpoints")
	}
}
