package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeFakeBin writes an executable that prints body and exits with code.
func makeFakeBin(t *testing.T, name, body string) string {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return d // return dir to prepend to PATH
}

// TestNewReportFile_CreatesAndCleans verifies the report file is created and
// the cleanup closure removes it.
func TestNewReportFile_CreatesAndCleans(t *testing.T) {
	p, cleanup, err := newReportFile()
	if err != nil {
		t.Fatalf("newReportFile: %v", err)
	}
	if _, e := os.Stat(p); e != nil {
		t.Fatalf("report file must exist: %v", e)
	}
	cleanup()
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("cleanup must remove the file")
	}
}

// TestCheckSecrets_FakeGitleaksFindsSecret exercises AC1 end-to-end via a
// fake gitleaks binary that writes a finding JSON to the report-path arg.
func TestCheckSecrets_FakeGitleaksFindsSecret(t *testing.T) {
	// The fake writes the finding JSON to $6 (detect --no-banner --report-format
	// json --report-path <path>; the path is the 6th positional arg).
	body := `printf '[{"RuleID":"fake-rule","File":"secret.go"}]' > "$6"
exit 1`
	dir := makeFakeBin(t, "gitleaks", body)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cfg := emptyPathCfg()
	r := checkSecrets(cfg, nil)
	if r.Status != Fail {
		t.Fatalf("fake gitleaks finding must yield Fail, got %v: %q", r.Status, r.Message)
	}
}

// TestCheckSecrets_FakeGitleaksClean exercises AC2 end-to-end via a fake
// gitleaks that writes an empty array to the report-path.
func TestCheckSecrets_FakeGitleaksClean(t *testing.T) {
	body := `printf '[]' > "$6"
exit 0`
	dir := makeFakeBin(t, "gitleaks", body)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cfg := emptyPathCfg()
	r := checkSecrets(cfg, nil)
	if r.Status != Pass {
		t.Fatalf("clean fake gitleaks must yield Pass, got %v: %q", r.Status, r.Message)
	}
}

// TestCheckSecrets_FakeGitleaksMalformedJSON exercises malformed JSON -> Warn.
func TestCheckSecrets_FakeGitleaksMalformedJSON(t *testing.T) {
	body := `printf 'NOT JSON' > "$6"
exit 1`
	dir := makeFakeBin(t, "gitleaks", body)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cfg := emptyPathCfg()
	r := checkSecrets(cfg, nil)
	if r.Status == Pass {
		t.Fatal("malformed gitleaks output must never yield Pass")
	}
	if r.Status != Warn {
		t.Fatalf("malformed output must yield Warn, got %v", r.Status)
	}
}

// TestVulnArgs_GovulncheckUsesJSONFlag verifies args for govulncheck.
func TestVulnArgs_GovulncheckUsesJSONFlag(t *testing.T) {
	args := vulnArgs("govulncheck")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-json") {
		t.Fatalf("govulncheck must use -json, got %v", args)
	}
}

// TestVulnArgs_OSVScannerUsesFormatJSON verifies args for osv-scanner.
func TestVulnArgs_OSVScannerUsesFormatJSON(t *testing.T) {
	args := vulnArgs("osv-scanner")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--format") || !strings.Contains(joined, "json") {
		t.Fatalf("osv-scanner must use --format json, got %v", args)
	}
}
