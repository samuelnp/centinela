package main

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

func buildStatusLineView(wfs []*workflow.Workflow) ui.StatusLineView {
	wf := primaryWorkflow(wfs)
	if wf == nil {
		return ui.StatusLineView{Primary: []string{"WF:none", "BLOCK:NO_WORKFLOW", "NEXT:start-feature"}}
	}
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	block, next := statusBlockAndNext(wf, cfg)
	mode := wf.OrchestrationMode
	if mode == "" {
		mode = "legacy"
	}
	progress := fmt.Sprintf("%d/%d", doneCount(wf), len(wf.OrderedSteps()))
	risk := "ok"
	if workflow.ProductionReadinessWarning(wf.Feature, cfg) != "" {
		risk = "warn"
	}
	return ui.StatusLineView{
		Primary: []string{"WF:" + wf.Feature, "STEP:" + wf.CurrentStep, "P:" + progress},
		Secondary: []string{
			"NEXT:" + next,
			"BLOCK:" + block,
			"MODE:" + mode,
			"RISK:" + risk,
		},
	}
}

func primaryWorkflow(wfs []*workflow.Workflow) *workflow.Workflow {
	for _, wf := range wfs {
		if wf.Feature == "" || wf.CurrentStep == "" || wf.CurrentStep == "done" || isRoleWorkflow(wf.Feature) {
			continue
		}
		return wf
	}
	for _, wf := range wfs {
		if wf.Feature != "" && wf.CurrentStep != "" && wf.CurrentStep != "done" {
			return wf
		}
	}
	return nil
}

func isRoleWorkflow(feature string) bool {
	return strings.HasSuffix(feature, "-big-thinker") ||
		strings.HasSuffix(feature, "-feature-specialist") ||
		strings.HasSuffix(feature, "-senior-engineer") ||
		strings.HasSuffix(feature, "-qa-senior")
}

func doneCount(wf *workflow.Workflow) int {
	if wf.CurrentStep == "done" {
		return len(wf.OrderedSteps())
	}
	for i, step := range wf.OrderedSteps() {
		if step == wf.CurrentStep {
			return i
		}
	}
	return 0
}
