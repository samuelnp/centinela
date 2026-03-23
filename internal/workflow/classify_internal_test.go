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
	if !IsAllowedInStep(TypePlan, "code") || !IsAllowedInStep(TypeCode, "tests") || !IsAllowedInStep(TypeOther, "validate") {
		t.Fatal("expected additional allowed combinations")
	}
	if IsAllowedInStep(TypeCode, "unknown") {
		t.Fatal("unknown step should be disallowed")
	}
	if isRoadmapArtifact("/p/other.md") {
		t.Fatal("non-roadmap file should not match")
	}
	if ClassifyFile("/p/tests/x.go", cfg) != TypeTests {
		t.Fatal("expected tests type")
	}
	if ClassifyFile("/p/misc/readme.txt", cfg) != TypeOther {
		t.Fatal("expected other type")
	}
}
