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
