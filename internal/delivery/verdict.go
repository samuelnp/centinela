package delivery

import "github.com/samuelnp/centinela/internal/gatereport"

// gatekeeperVerdict extracts the verdict from the report's "Status:" line ONLY
// (e.g. "**Status:** SAFE"), returning the first verdict token on that line.
// Delegated so there is ONE parser of the Status contract, not two; the
// composer surfaces the raw token, so legacy BLOCKING stays BLOCKING here.
func gatekeeperVerdict(report string) string {
	return gatereport.Verdict(report)
}
