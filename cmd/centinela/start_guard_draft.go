package main

import "fmt"

// draftStartError explains why `centinela start` refuses a draft feature and
// points the operator at the finalize path. The message contains "draft" so
// callers (and the acceptance suite) can assert the refusal reason.
func draftStartError(feature string) error {
	return fmt.Errorf(
		"cannot start %q — it is a draft feature with no quality scores yet; "+
			"finalize it first with centinela roadmap promote %s "+
			"--scores <ac,uv,dc,dep,ee,overall>",
		feature, feature)
}
