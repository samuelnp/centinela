package gates

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// The anti-weakening pin: nothing in this feature relaxes G1. A 101-line file
// with no justified exception is still a Fail, and the cap is still 100.
func TestCheckFileSize_CapAndSeverityUnchanged(t *testing.T) {
	if maxLines != 100 {
		t.Fatalf("the G1 cap must stay 100, got %d", maxLines)
	}
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("src", 0755)                                     //nolint:errcheck
	os.WriteFile("src/big.go", []byte(makeBigSource(101)), 0644) //nolint:errcheck
	r := checkFileSize(&config.Config{}, nil)
	if r.Status != Fail || len(r.Details) == 0 {
		t.Fatalf("a 101-line file must still Fail, got %+v", r)
	}
	if !strings.Contains(r.Message, "100 lines") {
		t.Fatalf("the fail message must still name the 100-line cap, got %q", r.Message)
	}
}

// A justified oversized file keeps its own distinct message naming the file,
// its line count and the justification (E15) — it is not the clean-pass text.
func TestCheckFileSize_JustifiedPassKeepsItsOwnMessage(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.MkdirAll("internal", 0755)                                      //nolint:errcheck
	os.WriteFile("internal/blob.go", []byte(makeBigSource(120)), 0644) //nolint:errcheck
	cfg := &config.Config{Gates: config.GatesConfig{FileSizeExceptions: []config.FileSizeException{
		{Path: "internal/blob.go", Kind: "configuration", Reason: "large static map", MaxLines: 130},
	}}}
	r := checkFileSize(cfg, nil)
	if r.Status != Pass {
		t.Fatalf("a justified oversized file must still pass, got %+v", r)
	}
	if !strings.Contains(r.Message, "justified exception") {
		t.Fatalf("justified pass must keep its distinct message, got %q", r.Message)
	}
	if len(r.Details) != 1 || !strings.Contains(r.Details[0], "120 lines") ||
		!strings.Contains(r.Details[0], "large static map") {
		t.Fatalf("details must name the file, line count and justification, got %v", r.Details)
	}
}

// R7: a Skip must never turn a green run red, and must not inflate the pass tally.
func TestAllPassed_SkipAndWarnDoNotFailTheRun(t *testing.T) {
	results := []Result{
		{Name: "G1: File Size", Status: Skip, Message: fileSizeNothingInspected},
		{Name: "G11: i18n", Status: Warn, Message: "1 locale configured"},
		{Name: "G2: Imports", Status: Pass, Message: "ok"},
	}
	if !AllPassed(results) {
		t.Fatal("Skip/Warn must not fail the run")
	}
	passes := 0
	for _, r := range results {
		if r.Status == Pass {
			passes++
		}
	}
	if passes != 1 {
		t.Fatalf("a skipped gate must not be counted as passed, got %d passes", passes)
	}
}
