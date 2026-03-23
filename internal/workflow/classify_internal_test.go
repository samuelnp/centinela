package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func TestClassifyAndAllowedInternal(t *testing.T) {
	cfg := &config.Config{}
	if ClassifyFile("/p/docs/plans/x.md", cfg) != TypePlan {
		t.Fatal("expected plan type")
	}
	cfg.Workflow.CodeDirs = []string{"/svc/"}
	if ClassifyFile("/p/svc/x.go", cfg) != TypeCode {
		t.Fatal("expected code type with custom dirs")
	}
	if !IsAllowedInStep(TypeTests, "tests") || IsAllowedInStep(TypeCode, "plan") {
		t.Fatal("unexpected IsAllowedInStep results")
	}
	if isRoadmapArtifact("/p/other.md") {
		t.Fatal("non-roadmap file should not match")
	}
}
