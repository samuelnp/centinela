package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// Scenario: verify_timeout below the suite floor is flagged as WARN
func TestDoctorConfigLowTimeoutWarn(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[verify]\nverify_timeout = 60\n")
	out, code := runDoctor(t, dir)
	if code != 0 || !strings.Contains(out, "⚠ config") {
		t.Fatalf("low timeout must WARN/exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "verify_timeout") || !strings.Contains(out, "180") {
		t.Fatalf("must mention verify_timeout + minimum:\n%s", out)
	}
}

// Scenario: Gate referencing a non-existent directory is flagged as WARN
func TestDoctorConfigMissingGateDirWarn(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml",
		"[verify]\nverify_timeout = 240\n[i18n]\ndir = \"missing-locales\"\n")
	out, code := runDoctor(t, dir)
	if code != 0 || !strings.Contains(out, "⚠ config") {
		t.Fatalf("missing gate dir must WARN/exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "missing-locales") {
		t.Fatalf("must name the missing directory:\n%s", out)
	}
}

// Scenario: Unknown TOML keys in centinela.toml are flagged as WARN
func TestDoctorConfigUnknownKeyWarn(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[verify]\nverify_timeout = 240\nbogus_key = 1\n")
	out, code := runDoctor(t, dir)
	if code != 0 || !strings.Contains(out, "⚠ config") {
		t.Fatalf("unknown key must WARN/exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "bogus_key") {
		t.Fatalf("must name the unknown key:\n%s", out)
	}
}

// Scenario: Config check is report-only and --fix does not modify centinela.toml
func TestDoctorConfigReportOnly(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "centinela.toml", "[verify]\nverify_timeout = 60\n")
	before, _ := os.ReadFile(filepath.Join(dir, "centinela.toml"))
	out, _ := runDoctor(t, dir, "--fix")
	after, _ := os.ReadFile(filepath.Join(dir, "centinela.toml"))
	if string(before) != string(after) {
		t.Fatal("--fix must NOT modify centinela.toml")
	}
	if !strings.Contains(out, "⚠ config") {
		t.Fatalf("config WARN must persist after --fix:\n%s", out)
	}
}
