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

// StewardPromptPath is the invocation guide the orchestrator must follow.
const StewardPromptPath = "docs/architecture/merge-steward-prompt.md"

// StewardDirective returns the two-line CENTINELA DIRECTIVE block that
// tells the orchestrator session to dispatch the merge-steward subagent
// and how to resume the merge afterwards. The wording mirrors the other
// directive hooks (imperative line + a details line).
func (o MergeOutcome) StewardDirective() string {
	return fmt.Sprintf(
		"CENTINELA DIRECTIVE: merge stalled (%s) for %q; delegate to merge-steward per %s.\n"+
			"Required evidence before resume: %s, %s. Then run: centinela merge --continue %s",
		o.StewardReason(), o.Feature, StewardPromptPath,
		o.StewardHint(), o.StewardJSONPath(), o.Feature,
	)
}
