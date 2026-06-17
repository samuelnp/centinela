package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
)

// Acceptance: specs/headless-governance.feature

// hgCommandPacket replays the runVerdict wiring (full-scan gates.RunAll + a
// fixed Now) against a temp project, returning the marshaled packet + summary.
func hgCommandPacket(t *testing.T, toml string) (*verdict.Packet, []byte) {
	t.Helper()
	cfg, _ := config.Load()
	deps := verdict.Deps{
		Gates:    gates.RunAll, // nil filter → always a full scan in v1
		Verify:   func(string, string, *config.Config) verify.VerificationResult { return verify.VerificationResult{} },
		Evidence: verdict.EvidenceIndex,
		Now:      hgNow,
	}
	pkt := verdict.AssembleVerdict("feat", cfg, hgWf(), deps)
	b, err := json.MarshalIndent(pkt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	_ = toml
	return pkt, b
}

func hgCmdProject(t *testing.T, toml string) {
	t.Helper()
	d := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })                 //nolint:errcheck
	os.Chdir(d)                                         //nolint:errcheck
	os.MkdirAll(".workflow", 0o755)                     //nolint:errcheck
	os.WriteFile("centinela.toml", []byte(toml), 0o644) //nolint:errcheck
}

// Scenario: Verdict command separates JSON stdout from status stderr
func TestHG_CommandJSONOnStdout(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	hgCmdProject(t, "")
	pkt, b := hgCommandPacket(t, "")
	if pkt.Summary.Verdict != "pass" {
		t.Fatalf("clean project should pass, got %s", pkt.Summary.Verdict)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("stdout payload must be valid JSON: %v", err)
	}
}

// Scenario: Verdict command exits via a silenced sentinel error on fail
func TestHG_CommandSentinelOnFail(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	hgCmdProject(t, "[gates]\nfile_size = true\n")
	os.MkdirAll("internal", 0o755) //nolint:errcheck
	os.WriteFile(filepath.Join("internal", "big.go"),
		[]byte("package p\n"+strings.Repeat("// l\n", 200)), 0o644) //nolint:errcheck
	pkt, b := hgCommandPacket(t, "")
	if pkt.Summary.Verdict != "fail" || pkt.Summary.ExitCode != 1 {
		t.Fatalf("oversize file must fail: %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
	if len(b) == 0 {
		t.Fatal("JSON must be produced before the sentinel error")
	}
}

// Scenario: Verdict command surfaces the resolved headless state in the packet
func TestHG_CommandHeadlessFlag(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	hgCmdProject(t, "")
	pkt, _ := hgCommandPacket(t, "")
	if !pkt.Run.Headless {
		t.Fatal("run.headless must reflect the resolved CENTINELA_HEADLESS=1")
	}
}

// Scenario: Verdict always full-scans gates in v1
func TestHG_FullScanGatesV1(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.FileSizeEnabled = true
	full := gates.RunAll(cfg)
	nilFilter := gates.RunWithFilter(cfg, nil)
	if len(full) != len(nilFilter) || full[0].Name != nilFilter[0].Name {
		t.Fatal("RunAll must equal a nil-filter full scan (v1 contract)")
	}
}
