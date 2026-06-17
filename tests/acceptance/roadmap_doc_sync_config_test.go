package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

// Scenario: An unknown severity value is rejected at config load
func TestRds_UnknownSeverityRejected(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap,
		"[gates]\nfile_size = false\n\n[gates.roadmap_drift]\nenabled = true\nseverity = \"error\"\n")
	out, code := rdsValidate(t, bin, dir)
	if code == 0 {
		t.Fatalf("bad severity must fail config load\n%s", out)
	}
	mustHave(t, out, "severity")
	if !strings.Contains(out, "fail") || !strings.Contains(out, "warn") {
		t.Fatalf("error must name valid values:\n%s", out)
	}
}

// Scenario: Unknown severity is a no-op when the gate is disabled
func TestRds_UnknownSeverityNoopWhenDisabled(t *testing.T) {
	bin := buildCent(t)
	dir := rdsDir(t, sampleRoadmap,
		"[gates]\nfile_size = false\n\n[gates.roadmap_drift]\nenabled = false\nseverity = \"bad\"\n")
	rdsGenerate(t, bin, dir)
	if _, code := rdsValidate(t, bin, dir); code != 0 {
		t.Fatal("disabled gate with bad severity must load and validate successfully")
	}
}

// Scenario: The drift gate is registered and ships with enabled true and severity warn
func TestRds_ShipsEnabledWarn(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "centinela.toml"))
	if err != nil {
		t.Fatal(err)
	}
	toml := string(data)
	idx := strings.Index(toml, "[gates.roadmap_drift]")
	if idx < 0 {
		t.Fatal("centinela.toml must register [gates.roadmap_drift]")
	}
	block := toml[idx:]
	if end := strings.Index(block[1:], "\n["); end >= 0 {
		block = block[:end+1]
	}
	mustHave(t, block, "enabled  = true")
	mustHave(t, block, "severity = \"warn\"")
}
