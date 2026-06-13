package ui

import (
	"strings"
	"testing"
)

func TestRenderChangelogNeeded(t *testing.T) {
	out := RenderChangelogNeeded("right-size-docs-step")
	if out == "" {
		t.Fatal("changelog-needed panel should render")
	}
	if !strings.Contains(out, "right-size-docs-step") {
		t.Fatal("panel must name the feature")
	}
	if !strings.Contains(out, "changelog") {
		t.Fatal("panel must mention the changelog")
	}
	if !strings.Contains(out, "artifact new right-size-docs-step changelog") {
		t.Fatal("panel must point at the artifact command")
	}
}
