package gates

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

const secretsName = "G-Secrets: Secret Scan"

// checkSecrets runs gitleaks over the working tree and maps the outcome to a
// single Result. Tool absent -> Skip; clean (or only allowlisted/out-of-diff
// findings) -> Pass; any retained finding -> Fail; a scan that ran but produced
// unusable output -> Warn (never a false Pass).
func checkSecrets(cfg *config.Config, filter *gitdiff.Set) Result {
	tool := cfg.Gates.Security.Secrets.Tool
	r := Result{Name: secretsName}
	if !toolPresent(tool) {
		r.Status = Skip
		r.Message = fmt.Sprintf("%s not installed; secret scan skipped.", tool)
		return r
	}
	report, cleanup, err := newReportFile()
	if err != nil {
		r.Status = Warn
		r.Message = "secrets: could not create report file: " + err.Error()
		return r
	}
	defer cleanup()
	_, stderr, runErr := runScanner(tool,
		"detect", "--no-banner", "--report-format", "json", "--report-path", report)
	if runErr == errScanTimeout {
		r.Status = Warn
		r.Message = "secrets: gitleaks timed out."
		return r
	}
	return classifySecrets(report, string(stderr), runErr, cfg.Gates.Security.Secrets, filter)
}

// classifySecrets reads the gitleaks JSON report and folds it into a Result.
// gitleaks exits 0 with an empty report when clean and non-zero when findings
// exist; either way the report file is the source of truth. A launch failure
// (exit code -1) with no readable report is a tool error -> Warn.
func classifySecrets(report, stderr string, runErr error, cfg config.SecretsConfig, filter *gitdiff.Set) Result {
	r := Result{Name: secretsName}
	findings, perr := readSecretsReport(report)
	if perr != nil {
		r.Status = Warn
		r.Message = "secrets: gitleaks output unusable: " + perr.Error()
		return r
	}
	if findings == nil && runErr != nil && exitCode(runErr) < 0 {
		r.Status = Warn
		r.Message = "secrets: gitleaks failed to run: " + firstStderrLine(stderr, runErr)
		return r
	}
	kept := retainFindings(findings, cfg.Allowlist, filter)
	if len(kept) == 0 {
		r.Status = Pass
		r.Message = "No secrets detected."
		return r
	}
	r.Status = Fail
	r.Message = "Secrets detected — remove or allowlist before validate passes:"
	r.Details = kept
	return r
}

// newReportFile creates a temp file for gitleaks JSON output and returns its
// path plus a cleanup closure. The file is created then closed so gitleaks can
// own it; cleanup removes it.
func newReportFile() (string, func(), error) {
	f, err := os.CreateTemp("", "centinela-gitleaks-*.json")
	if err != nil {
		return "", func() {}, err
	}
	path := f.Name()
	_ = f.Close()
	return path, func() { _ = os.Remove(path) }, nil
}
