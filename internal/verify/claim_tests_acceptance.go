package verify

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/acceptance"
)

// runVerdict is one command's classification plus the advisory note a PASS
// carried (a parse limitation, or skip counts the policy chose not to fail on).
// The note is a FIELD, not something a caller re-parses out of Detail.
type runVerdict struct {
	check Check
	note  string
}

// classifyAcceptanceRun judges an exit-0 run. A non-acceptance command is a
// plain pass: classification gates the parse and the parse gates the verdict,
// so a legitimately skipping unit or integration tier can never be failed here.
//
// An UNPARSEABLE report stays StatusPass with the limitation named in the note.
// The verifier must not invent a failure it cannot prove — an undetected skip
// is not a proven skip.
func classifyAcceptanceRun(c Check, cmd, output string, isAcceptance bool, policy string) runVerdict {
	c.Status = StatusPass
	c.Detail = fmt.Sprintf("command %q exited 0", cmd)
	if !isAcceptance {
		return runVerdict{check: c}
	}
	verdict, detail := acceptance.Judge(cmd, output, policy)
	if verdict == acceptance.VerdictFail {
		c.Status = StatusFail
		c.Detail = fmt.Sprintf("acceptance skips detected — %s", detail)
		return runVerdict{check: c}
	}
	if detail != "" {
		c.Detail = fmt.Sprintf("command %q exited 0 — %s", cmd, detail)
	}
	return runVerdict{check: c, note: detail}
}
