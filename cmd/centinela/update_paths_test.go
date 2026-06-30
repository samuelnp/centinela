package main

import (
	"os"
	"strings"
	"testing"
)

// withExitCapture overrides exitMain to record the code instead of exiting.
func withExitCapture(t *testing.T) *int {
	t.Helper()
	var code int
	orig := exitMain
	exitMain = func(c int) { code = c }
	t.Cleanup(func() { exitMain = orig })
	return &code
}

func TestRunUpdateInstalls(t *testing.T) {
	bin, done := fakeUpdater(t, "0.37.0", 0)
	defer done()
	updateCheck = false
	out := capture(t, func() error { return runUpdate(updateCmd, nil) })
	if !strings.Contains(out, "0.37.0 -> 0.40.2") {
		t.Fatalf("out = %q", out)
	}
	if got, _ := os.ReadFile(bin); string(got) != "NEW" {
		t.Fatalf("binary not replaced: %q", got)
	}
}

func TestRunUpdateCheckBehindExits1(t *testing.T) {
	_, done := fakeUpdater(t, "0.37.0", 0)
	defer done()
	updateCheck = true
	defer func() { updateCheck = false }()
	code := withExitCapture(t)
	out := capture(t, func() error { return runUpdate(updateCmd, nil) })
	if !strings.Contains(out, "update available") || *code != 1 {
		t.Fatalf("out=%q code=%d", out, *code)
	}
}

func TestRunUpdateCheckCurrentExits0(t *testing.T) {
	_, done := fakeUpdater(t, "0.40.2", 0)
	defer done()
	updateCheck = true
	defer func() { updateCheck = false }()
	code := withExitCapture(t)
	out := capture(t, func() error { return runUpdate(updateCmd, nil) })
	if !strings.Contains(out, "up to date") || *code != 0 {
		t.Fatalf("out=%q code=%d", out, *code)
	}
}

func TestRunUpdateAPIErrorReturned(t *testing.T) {
	_, done := fakeUpdater(t, "0.37.0", 403)
	defer done()
	updateCheck = false
	err := runUpdate(updateCmd, nil)
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestEmitUpdateNoticePrints(t *testing.T) {
	_, done := fakeUpdater(t, "0.37.0", 0)
	defer done()
	out := capture(t, func() error { emitUpdateNotice(); return nil })
	if !strings.Contains(out, "update available") {
		t.Fatalf("notice out = %q", out)
	}
}
