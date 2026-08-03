package workflow

import "encoding/json"

// UnknownStep is the current step reported for a state file that came from a
// newer Centinela and whose own step this binary could not read. It is not a
// real step: nothing advances to it and no step rule matches it, which is why
// EvaluatePrewrite consults Unmodellable rather than the step name.
const UnknownStep = "unknown"

// stateMarkers are the keys that identify a document under .workflow/ as
// feature state. A future-versioned file carrying none of them is not
// recognisable as a workflow by this binary at all — and other .workflow/ JSON
// (roadmap.json, audit-baseline.json) carries none of them, which is what stops
// a future roadmap schema from being mistaken for a phantom workflow.
var stateMarkers = [...]string{"feature", "currentStep", "steps", "stepOrder"}

// degradedWorkflow is what Load returns for a state file whose schemaVersion is
// higher than this binary's, or unreadable, AND whose body this binary cannot
// unmarshal (a future version that changed a TYPE or a SHAPE rather than only
// adding fields).
//
// The contract, chosen so that a newer file can never brick an older binary:
//
//   - Load SUCCEEDS. It never returns an error for a future-versioned file, so
//     ActiveWorkflows keeps the feature, the active set is never emptied,
//     EvaluatePrewrite never falls to NeedInit and hook_autostart never forks a
//     duplicate workflow.
//   - The returned workflow is marked Unmodellable. EvaluatePrewrite ALLOWS
//     every governed write against it: this binary cannot enforce step rules
//     whose schema it does not understand, and refusing to govern is far
//     cheaper than blocking every write in the repo.
//   - Feature and CurrentStep are salvaged when they are still plain strings,
//     so a mis-named or per-role JSON is still filtered out by ActiveWorkflows;
//     otherwise the feature is the file's own name and the step is UnknownStep.
//   - SchemaVersion is the probed version, or 0 when it could not be read.
//   - Save REFUSES (the version guard re-reads the file), so nothing this
//     binary does can truncate the parts of the file it cannot model.
//
// nil means "this is not a future state file, or not state at all": the caller
// reports the parse error instead.
func degradedWorkflow(feature string, raw []byte, version int, readable bool) *Workflow {
	if readable && version <= SchemaVersion {
		return nil
	}
	fields := stateFields(raw)
	if !hasStateMarker(fields) {
		return nil
	}
	wf := &Workflow{
		SchemaVersion: version,
		Feature:       feature,
		CurrentStep:   UnknownStep,
		Steps:         map[string]StepState{},
		unmodellable:  true,
		loadedDigest:  digestOf(raw),
	}
	if s := rawString(fields["feature"]); s != "" {
		wf.Feature = s
	}
	if s := rawString(fields["currentStep"]); s != "" {
		wf.CurrentStep = s
	}
	return wf
}

func hasStateMarker(fields map[string]json.RawMessage) bool {
	for _, k := range stateMarkers {
		if _, ok := fields[k]; ok {
			return true
		}
	}
	return false
}

func rawString(token json.RawMessage) string {
	var s string
	if err := json.Unmarshal(token, &s); err != nil {
		return ""
	}
	return s
}

// Unmodellable reports whether this workflow was loaded from a state file whose
// schema this binary cannot parse (see degradedWorkflow). Governance decisions
// must not be enforced against it — only reported.
func (wf *Workflow) Unmodellable() bool { return wf.unmodellable }
