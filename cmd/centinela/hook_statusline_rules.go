package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func statusBlockAndNext(wf *workflow.Workflow, cfg *config.Config) (string, string) {
	if wf.CurrentStep == "plan" {
		if !fileExists("docs/features/" + wf.Feature + ".md") {
			return "MISSING_FEATURE_BRIEF", "write-brief"
		}
		if !fileExists("docs/plans/" + wf.Feature + ".md") {
			return "MISSING_PLAN", "write-plan"
		}
		if len(specFiles()) == 0 {
			return "MISSING_SPEC", "write-spec"
		}
		return "none", "complete-step"
	}
	if wf.CurrentStep == "tests" && !fileExists(".workflow/"+wf.Feature+"-edge-cases.md") {
		return "MISSING_EDGE_CASES", "write-edge-cases"
	}
	if wf.CurrentStep == "validate" {
		if !fileExists(".workflow/" + wf.Feature + "-gatekeeper.md") {
			return "MISSING_GATEKEEPER", "run-gatekeeper"
		}
		if err := workflow.ValidateArtifacts(wf.Feature, "validate", cfg); err != nil {
			if strings.Contains(err.Error(), "BLOCKING") {
				return "PROD_BLOCKING", "harden-feature"
			}
			return "MISSING_PROD_READINESS", "run-production-readiness"
		}
		return "none", "run-validate"
	}
	if wf.CurrentStep == "docs" {
		if orchestration.IsUserFacingFeature(wf.Feature) {
			if !fileExists("docs/project-docs/index.html") {
				return "MISSING_DOCS_OUTPUT", "run-documentation-specialist"
			}
		} else if !fileExists(".workflow/" + wf.Feature + "-changelog.md") {
			return "MISSING_DOCS_OUTPUT", "write-changelog"
		}
		if err := workflow.ValidateArtifacts(wf.Feature, "docs", cfg); err != nil {
			return "MISSING_DOCS_EVIDENCE", "write-docs-evidence"
		}
		return "none", "complete-step"
	}
	return "none", "implement-" + wf.CurrentStep
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func specFiles() []string {
	files, _ := filepath.Glob("specs/*.feature")
	return files
}
