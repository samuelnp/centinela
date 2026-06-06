package gates

import (
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// checkSecurity runs the enabled security checks and returns one Result per
// concern: G-Secrets (gitleaks, diff-aware locally, hard Fail) and G-Vuln
// (govulncheck + osv-scanner, whole-project, Warn-only). When every scanner
// family is absent both Skips carry a distinct "no scanners available" message
// so the run is not mistaken for a verified clean scan.
func checkSecurity(cfg *config.Config, filter *gitdiff.Set) []Result {
	secrets := checkSecrets(cfg, filter)
	vuln := checkVuln(cfg)
	if secrets.Status == Skip && vuln.Status == Skip {
		const note = " No security scanners available — nothing was verified."
		secrets.Message += note
		vuln.Message += note
	}
	return []Result{secrets, vuln}
}
