package main

import (
	"os"
	"testing"
)

// corruptConfigDir chdirs into a temp dir whose centinela.toml is unparseable,
// so config.Load() surfaces an error every consumer must propagate.
func corruptConfigDir(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.WriteFile("centinela.toml", []byte("[cost\nenabled = true"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunCalibrateConfigLoadError(t *testing.T) {
	corruptConfigDir(t)
	if err := runCalibrate(nil, nil); err == nil {
		t.Fatal("corrupt config should surface from runCalibrate")
	}
}

func TestRunCostConfigLoadError(t *testing.T) {
	corruptConfigDir(t)
	if err := runCost(nil, nil); err == nil {
		t.Fatal("corrupt config should surface from runCost")
	}
}

func TestRunVerdictConfigLoadError(t *testing.T) {
	corruptConfigDir(t)
	if err := runVerdict(nil, []string{"feat"}); err == nil {
		t.Fatal("corrupt config should surface from runVerdict")
	}
}

func TestRunVerifyConfigLoadError(t *testing.T) {
	corruptConfigDir(t)
	if err := runVerify(nil, []string{"feat"}); err == nil {
		t.Fatal("corrupt config should surface from runVerify")
	}
}
