package integration_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestClassifyFile_DocsFeatures_IsRoadmapNotPlan(t *testing.T) {
	cfg := &config.Config{}
	got := workflow.ClassifyFile("/project/docs/features/my-feature.md", cfg)
	if got == workflow.TypePlan {
		t.Error("docs/features/ should be TypeRoadmap, not TypePlan")
	}
	if got != workflow.TypeRoadmap {
		t.Errorf("expected TypeRoadmap, got %q", got)
	}
}

func TestClassifyFile_DocsPlans_IsStillPlan(t *testing.T) {
	cfg := &config.Config{}
	got := workflow.ClassifyFile("/project/docs/plans/my-feature.md", cfg)
	if got != workflow.TypePlan {
		t.Errorf("docs/plans/ should be TypePlan, got %q", got)
	}
}
