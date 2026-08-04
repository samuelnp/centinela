package hookpolicy

import (
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
	// StaleBinary blocks because every active workflow's state file was written
	// by a NEWER Centinela than this one, so no step rule here can be trusted.
	// It is a refusal, not a pass: `.workflow/*.json` is an ungoverned write
	// target, so allowing here would let an agent unblock itself by corrupting
	// its own state file. The remedy is to upgrade the binary, not to guess.
	StaleBinary bool
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
	active, degraded := 0, 0
	for _, wf := range wfs {
		if wf.CurrentStep == "done" {
			continue
		}
		active++
		// A workflow whose state file came from a NEWER Centinela cannot be
		// modelled here, so it is never ENFORCED. It is not allowed to widen
		// permissions either: returning Allow for it would let one unreadable
		// file disarm step gating for every other feature in the repo. It only
		// yields when it is the sole reason to say no (see below).
		if wf.Unmodellable() {
			degraded++
			continue
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
	// Every active workflow is unmodellable. Refuse: this binary cannot read the
	// step rules, and `.workflow/*.json` is itself an ungoverned write target,
	// so passing here would hand any agent a self-service bypass — write a
	// future version over your own state file and the gate opens. The refusal
	// names the real remedy instead, and Load still succeeded, so the workflow
	// is not forgotten and no duplicate is auto-started.
	if degraded == active {
		return PrewriteDecision{FileType: fileType, StaleBinary: true, Feature: firstDegraded(wfs).Feature}
	}
	wf := firstEnforceable(wfs)
	return PrewriteDecision{FileType: fileType, Step: wf.CurrentStep, Feature: wf.Feature}
}
