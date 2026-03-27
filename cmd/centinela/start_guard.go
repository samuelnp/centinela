package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/projectstage"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func workflowOrderForFeature(feature string) ([]string, error) {
	stage, err := projectstage.Load("PROJECT.md")
	if err != nil {
		return nil, err
	}
	if stage == projectstage.Existing {
		return workflow.DefaultStepOrder, nil
	}
	r, err := roadmap.Load()
	if err != nil {
		return nil, fmt.Errorf("greenfield project requires roadmap before start")
	}
	if err := roadmap.ValidateAnalysis(r); err != nil {
		return nil, fmt.Errorf("greenfield project requires roadmap senior PM analysis: %w", err)
	}
	if !roadmap.HasBootstrapPhase(r) {
		return nil, fmt.Errorf("greenfield project requires roadmap phase \"Phase 0: Bootstrap\"")
	}
	if roadmap.IsBootstrapFeature(r, feature) {
		return workflow.BootstrapStepOrder, nil
	}
	if !roadmap.BootstrapComplete(r) {
		return nil, fmt.Errorf("bootstrap is incomplete; start a Phase 0 feature first")
	}
	return workflow.DefaultStepOrder, nil
}
