package gates

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// vulnCfg returns a security config whose only vuln tool is govulncheck, so a
// single fake binary on PATH controls the whole gate outcome.
func vulnCfg() *config.Config {
	c := &config.Config{}
	c.Gates.Security.Enabled = true
	c.Gates.Security.Vuln.Tools = []string{"govulncheck"}
	return c
}

// onlyFakeBin makes PATH contain ONLY the dir holding the fake binary so every
// other configured tool deterministically resolves as absent.
func onlyFakeBin(t *testing.T, name, body string) {
	t.Helper()
	dir := makeFakeBin(t, name, body)
	t.Setenv("PATH", dir)
}

// TestCheckVuln_FakeFindingYieldsWarnWithDetails drives runVulnTool + foldVuln
// through the findings arm via a fake govulncheck emitting one NDJSON finding.
func TestCheckVuln_FakeFindingYieldsWarnWithDetails(t *testing.T) {
	onlyFakeBin(t, "govulncheck",
		`printf '{"finding":{"osv":"GO-2024-0001","trace":[{"module":"example.com/x"}]}}\n'
exit 1`)
	r := checkVuln(vulnCfg())
	if r.Status != Warn {
		t.Fatalf("a finding must yield Warn, got %v: %q", r.Status, r.Message)
	}
	if len(r.Details) != 1 || r.Details[0] != "example.com/x: GO-2024-0001" {
		t.Fatalf("unexpected details: %v", r.Details)
	}
}

// TestCheckVuln_FakeCleanYieldsPass drives the clean arm: present tool, empty
// output, exit 0 -> Pass.
func TestCheckVuln_FakeCleanYieldsPass(t *testing.T) {
	onlyFakeBin(t, "govulncheck", "exit 0")
	r := checkVuln(vulnCfg())
	if r.Status != Pass {
		t.Fatalf("clean scan must yield Pass, got %v: %q", r.Status, r.Message)
	}
}

// TestCheckVuln_FakeMalformedYieldsIncompleteWarn drives runVulnTool's parse
// error -> note -> foldVuln's warns arm ("Dependency audit incomplete").
func TestCheckVuln_FakeMalformedYieldsIncompleteWarn(t *testing.T) {
	onlyFakeBin(t, "govulncheck", `printf '{not json'
exit 0`)
	r := checkVuln(vulnCfg())
	if r.Status != Warn {
		t.Fatalf("malformed output must yield Warn, got %v", r.Status)
	}
	if len(r.Details) == 0 {
		t.Fatalf("incomplete audit must carry a note, got none")
	}
}

// TestRunVulnTool_FakeFindingReturnsKeys exercises runVulnTool directly: it must
// surface the parsed key and no note when the fake scanner output is valid.
func TestRunVulnTool_FakeFindingReturnsKeys(t *testing.T) {
	onlyFakeBin(t, "govulncheck",
		`printf '{"finding":{"osv":"GO-1","trace":[{"module":"m"}]}}\n'`)
	keys, note := runVulnTool("govulncheck")
	if note != "" {
		t.Fatalf("valid output must produce no note, got %q", note)
	}
	if len(keys) != 1 || keys[0].ID != "GO-1" {
		t.Fatalf("expected one parsed key, got %v", keys)
	}
}
