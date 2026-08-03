package hookpolicy

import (
	"path/filepath"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// PrewriteDecision is the result of evaluating a pending write.
type PrewriteDecision struct {
	Allow    bool
	NeedInit bool
	FileType workflow.FileType
	Step     string
	Feature  string
	// Path is the offending path on a blocking decision. EvaluatePrewriteMulti
	// sets it so callers can render which of several patched files was blocked.
	Path string
}

// EvaluatePrewrite decides whether a file write is allowed.
func EvaluatePrewrite(path, cwd string, cfg *config.Config, wfs []*workflow.Workflow) PrewriteDecision {
	if !isInsideWorkspace(path, cwd) {
		return PrewriteDecision{Allow: true}
	}
	fileType := workflow.ClassifyFile(path, cfg)
	if fileType == workflow.TypeOther || fileType == workflow.TypeRoadmap {
		return PrewriteDecision{Allow: true, FileType: fileType}
	}
	if len(wfs) == 0 {
		return PrewriteDecision{FileType: fileType, NeedInit: true}
	}
	active := 0
	for _, wf := range wfs {
		if wf.CurrentStep == "done" {
			continue
		}
		active++
		// A workflow whose state file came from a NEWER Centinela and cannot be
		// modelled here is reported, never enforced: this binary does not
		// understand that file's step semantics, and guessing them would block
		// every governed write in the repo (see workflow.Unmodellable).
		if wf.Unmodellable() {
			return PrewriteDecision{Allow: true, FileType: fileType}
		}
		// outcome drops the ordering rails: any write in an active step is
		// allowed. The no-active-workflow block above is untouched for all
		// profiles. strict/guided keep today's step-gating.
		if workflow.EffectiveProfile(wf, cfg) == config.ProfileOutcome ||
			workflow.IsAllowedInStep(fileType, wf.CurrentStep) {
			return PrewriteDecision{Allow: true, FileType: fileType}
		}
	}
	if active == 0 {
		return PrewriteDecision{FileType: fileType, NeedInit: true}
	}
	wf := firstActive(wfs)
	return PrewriteDecision{FileType: fileType, Step: wf.CurrentStep, Feature: wf.Feature}
}

func firstActive(wfs []*workflow.Workflow) *workflow.Workflow {
	for _, wf := range wfs {
		if wf.CurrentStep != "done" {
			return wf
		}
	}
	return &workflow.Workflow{}
}

func isInsideWorkspace(path, cwd string) bool {
	if cwd == "" {
		return true
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
