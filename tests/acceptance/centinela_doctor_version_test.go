package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runDoctorPATH runs the doctor binary in dir with a controlled PATH so the
// version check observes a stubbed (or absent) `centinela` binary.
func runDoctorPATH(t *testing.T, dir, pathEnv string, args ...string) (string, int) {
	t.Helper()
	bin := buildDoctorBin(t)
	c := exec.Command(bin, append([]string{"doctor"}, args...)...)
	c.Dir = dir
	c.Env = append(os.Environ(), "PATH="+pathEnv)
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	return string(out), code
}

// fakeCentinela writes a stub `centinela` reporting ver and returns its bin dir.
func fakeCentinela(t *testing.T, ver string) string {
	t.Helper()
	bindir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bindir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho \"centinela version " + ver + "\"\n"
	if err := os.WriteFile(filepath.Join(bindir, "centinela"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bindir
}

// Acceptance: specs/centinela-doctor.feature

// Scenario: Installed binary version behind Makefile VERSION is flagged as WARN
func TestDoctorVersionBehindWarn(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "Makefile", "VERSION := 0.21.1\n")
	bindir := fakeCentinela(t, "0.15.0")
	out, code := runDoctorPATH(t, dir, bindir)
	if code != 0 || !strings.Contains(out, "⚠ version") {
		t.Fatalf("behind must WARN/exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "0.15.0") || !strings.Contains(out, "0.21.1") {
		t.Fatalf("must report both versions:\n%s", out)
	}
	if !strings.Contains(out, "make install") {
		t.Fatalf("must recommend make install:\n%s", out)
	}
}

// Scenario: Binary version check is report-only and --fix does not reinstall
func TestDoctorVersionReportOnly(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "Makefile", "VERSION := 0.21.1\n")
	bindir := fakeCentinela(t, "0.15.0")
	out, _ := runDoctorPATH(t, dir, bindir, "--fix")
	if !strings.Contains(out, "⚠ version") || !strings.Contains(out, "make install") {
		t.Fatalf("version WARN + make install must persist after --fix:\n%s", out)
	}
}

// Scenario: centinela binary not found on PATH causes version check to degrade to WARN
func TestDoctorVersionBinaryNotFoundWarn(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, "Makefile", "VERSION := 0.21.1\n")
	emptyBin := filepath.Join(t.TempDir(), "empty")
	_ = os.MkdirAll(emptyBin, 0o755)
	out, code := runDoctorPATH(t, dir, emptyBin)
	if code != 0 || !strings.Contains(out, "⚠ version") {
		t.Fatalf("missing binary must WARN/exit 0, got %d\n%s", code, out)
	}
}
