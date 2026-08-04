package main

import "fmt"

// draftStartError explains why `centinela start` refuses a draft feature and
// points the operator at the finalize path.
//
// The refusal stands on its own merits and always did: a draft is a feature
// whose phase, definition and dependencies nobody has agreed to yet, so there
// is nothing for the workflow to be a workflow OF. It never depended on the
// self-graded overall >= 9 gate, which is now deleted — finalizing records
// scores, it does not clear a bar. The message contains "draft" so callers
// (and the acceptance suite) can assert the refusal reason.
func draftStartError(feature string) error {
	return fmt.Errorf(
		"cannot start %q — it is a draft feature: its phase and definition are "+
			"not agreed yet; finalize it first with centinela roadmap promote %s "+
			"--scores <ac,uv,dc,dep,ee,overall>",
		feature, feature)
}
