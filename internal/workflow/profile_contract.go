package workflow

// ProfileContractGuidedDefault is the profile-default contract pinned by
// NewWithOrder on every workflow created after the guided-by-default flip.
//
// Back-compat here is STATE-DATED, never clock-dated, exactly like
// ValidateContract and PlanContract: the tail (nothing-configured) tier of
// EffectiveProfile returns guided only for a workflow that carries this pin.
// A workflow started before the flip — or one already in flight — has an empty
// ProfileContract and keeps resolving to strict for its whole life, so the
// default change can never loosen enforcement mid-run. The pin is written at
// start and never rewritten, so a fresh workflow cannot dodge the new default
// by hand-editing state either: removing the pin only makes it stricter.
const ProfileContractGuidedDefault = "guided-default-v1"

// UsesGuidedDefault reports whether this workflow was created under the
// guided-by-default contract and may therefore take guided as its tail default.
// A nil workflow and an unpinned (legacy or in-flight) workflow both report
// false, which is what keeps their resolution strict.
func (wf *Workflow) UsesGuidedDefault() bool {
	return wf != nil && wf.ProfileContract == ProfileContractGuidedDefault
}
