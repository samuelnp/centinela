package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRenderStatusRevisionsRow(t *testing.T) {
	wf := workflow.New("f")
	wf.Revisions = []workflow.Revision{{From: "validate", To: "code", Reason: "bug found"}}
	out := RenderStatus(wf)
	if !strings.Contains(out, "Revisions") {
		t.Fatal("expected a Revisions row")
	}
	if !strings.Contains(out, "bug found") {
		t.Fatal("expected the latest reason inline")
	}
}

func TestRenderStatusNoRevisionsRow(t *testing.T) {
	out := RenderStatus(workflow.New("f"))
	if strings.Contains(out, "Revisions") {
		t.Fatal("Revisions row must be omitted when there are none")
	}
}
