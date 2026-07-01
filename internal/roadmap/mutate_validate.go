package roadmap

import "fmt"

// requirePlannedStatus rejects a mutation on a feature whose derived status is
// in-progress or done. Only planned features may be removed; an active feature
// carries workflow state that a raw delete would orphan. The status string is
// surfaced verbatim so the caller sees "in-progress"/"done" in the message.
func requirePlannedStatus(slug string) error {
	if st := FeatureStatus(slug); st != "planned" {
		return fmt.Errorf(
			"cannot remove %q: its status is %s (only planned features can be removed)",
			slug, st)
	}
	return nil
}

// requireNoDependents rejects a mutation on a feature that other features still
// depend on, naming the dependents (drafts included) so the operator can act.
func (d *rawDoc) requireNoDependents(slug string) error {
	deps, err := d.featureDependents(slug)
	if err != nil {
		return err
	}
	if len(deps) > 0 {
		return fmt.Errorf(
			"cannot remove %q: it is depended on by %s", slug, joinNames(deps))
	}
	return nil
}

// joinNames renders a comma-separated, quoted name list for error messages.
func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%q", n)
	}
	return out
}
