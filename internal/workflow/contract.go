package workflow

// ValidateContractAdversarial is the validate-step contract pinned on every
// workflow started after the adversarial verifier shipped. Back-compat is
// state-dated, never clock-dated (D4): an empty ValidateContract means the
// workflow predates this contract and keeps the old existence-only gate, and a
// fresh feature cannot dodge the new format by hand-authoring legacy files.
const ValidateContractAdversarial = "adversarial-v1"

// UsesAdversarialVerifier reports whether the feature's validate step is
// governed by the grounded-verdict gate. A nil or unpinned workflow is legacy.
func (wf *Workflow) UsesAdversarialVerifier() bool {
	return wf != nil && wf.ValidateContract == ValidateContractAdversarial
}

// featureUsesAdversarialVerifier resolves the pinned contract from disk. An
// unreadable workflow is treated as legacy so a state-file problem never
// invents a gate the operator cannot satisfy.
func featureUsesAdversarialVerifier(feature string) bool {
	wf, err := Load(feature)
	if err != nil {
		return false
	}
	return wf.UsesAdversarialVerifier()
}
