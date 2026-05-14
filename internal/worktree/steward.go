package worktree

import "fmt"

// StewardHint returns the .workflow path where the Merge Steward will write
// its evidence for this outcome's feature. Used to guide the operator.
func (o MergeOutcome) StewardHint() string {
	return fmt.Sprintf(".workflow/%s-merge-steward.md", o.Feature)
}

// StewardJSONPath returns the structured evidence path for the feature.
func (o MergeOutcome) StewardJSONPath() string {
	return fmt.Sprintf(".workflow/%s-merge-steward.json", o.Feature)
}

// StewardReason categorises why the steward needs to be invoked.
func (o MergeOutcome) StewardReason() string {
	switch {
	case o.TextConflict:
		return "git-text-conflict"
	case o.ValidateFail:
		return "post-merge-validate-failed"
	default:
		return ""
	}
}
