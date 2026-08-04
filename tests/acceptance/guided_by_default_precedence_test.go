// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: An explicit per-feature profile still outranks the new default
func TestGBD_ExplicitFlagOutranksDefault(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdExisting(t, dir)
	if out, code := runCent(t, bin, dir, "start", "flagged", "--profile", "strict"); code != 0 {
		t.Fatalf("start failed: %s", out)
	}
	out := gbdStatus(t, bin, dir, "flagged")
	if !strings.Contains(out, "strict") || !strings.Contains(out, "--profile") {
		t.Fatalf("--profile strict must outrank the guided default, got: %s", out)
	}
}

// Scenario: An explicit global profile still outranks the new default
func TestGBD_ExplicitGlobalOutranksDefault(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdExisting(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[workflow]\nenforcement_profile = \"strict\"\n")
	if out, code := runCent(t, bin, dir, "start", "globaled"); code != 0 {
		t.Fatalf("start failed: %s", out)
	}
	out := gbdStatus(t, bin, dir, "globaled")
	if !strings.Contains(out, "strict") || !strings.Contains(out, "global") {
		t.Fatalf("an explicit global strict must outrank the guided default, got: %s", out)
	}
}

// Scenario: A driver model's capability class still outranks the new default
func TestGBD_LimitedDriverModelOutranksDefault(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdExisting(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[orchestration]\ndriver_model = \"haiku\"\n")
	if out, code := runCent(t, bin, dir, "start", "limited-driver"); code != 0 {
		t.Fatalf("start failed: %s", out)
	}
	out := gbdStatus(t, bin, dir, "limited-driver")
	if !strings.Contains(out, "strict") || !strings.Contains(out, "driver: haiku") {
		t.Fatalf("a limited driver model must still resolve to strict, got: %s", out)
	}
}

// Scenario: A capability-derived guided profile is distinguishable from the default
func TestGBD_CapableDriverModelDistinctFromDefault(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	gbdExisting(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), "[orchestration]\ndriver_model = \"sonnet\"\n")
	if out, code := runCent(t, bin, dir, "start", "capable-driver"); code != 0 {
		t.Fatalf("start failed: %s", out)
	}
	out := gbdStatus(t, bin, dir, "capable-driver")
	if !strings.Contains(out, "guided") || !strings.Contains(out, "driver: sonnet") {
		t.Fatalf("a capable driver model's guided must name the driver, got: %s", out)
	}
	if strings.Contains(out, "default (guided)") {
		t.Fatalf("capability-derived guided must not read as the bare default: %s", out)
	}
}
