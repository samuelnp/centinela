package config

import (
	"fmt"
	"strings"
)

// [validate] acceptance_skip_policy decides what an acceptance-classified
// command that exited 0 while reporting skipped / pending / undefined scenarios
// does to the run. It defaults ON: a suite that asserted nothing is not a pass.
// The same three literals are restated in the stdlib-only internal/acceptance
// leaf, which cannot import this package.
const (
	AcceptanceSkipFail = "fail"
	AcceptanceSkipWarn = "warn"
	AcceptanceSkipOff  = "off"
)

// NormalizeAcceptanceSkipPolicy maps any input to a known policy. Empty falls
// through to fail — the default-on behavior.
func NormalizeAcceptanceSkipPolicy(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case AcceptanceSkipWarn:
		return AcceptanceSkipWarn
	case AcceptanceSkipOff:
		return AcceptanceSkipOff
	default:
		return AcceptanceSkipFail
	}
}

// validateAcceptanceSkipPolicy rejects a NON-EMPTY unsupported value. It must
// run against the RAW decoded value, before applyDefaults normalizes it — the
// normalizer would otherwise silently swallow the operator's typo.
func validateAcceptanceSkipPolicy(policy string) error {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "", AcceptanceSkipFail, AcceptanceSkipWarn, AcceptanceSkipOff:
		return nil
	default:
		return fmt.Errorf("validate.acceptance_skip_policy %q is unsupported (use fail, warn, or off)", policy)
	}
}
