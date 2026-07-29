package worktree

// Directive re-renders the dispatch directive from a stored marker so the
// UserPromptSubmit hook can re-emit it verbatim while the marker lives.
func (m *PendingMarker) Directive() string {
	o := MergeOutcome{Feature: m.Feature, ConflictedPaths: m.ConflictedPaths}
	switch m.Reason {
	case "git-text-conflict":
		o.TextConflict = true
	case "post-merge-validate-failed":
		o.ValidateFail = true
	}
	return o.StewardDirective()
}
