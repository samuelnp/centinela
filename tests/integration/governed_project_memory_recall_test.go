package integration_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/memory"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

// SC-08/10: plan advisor recall path runs without error over a populated ledger.
func TestMemoryIntegration_RecallInPlanAdvisor(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0755)     //nolint:errcheck
	os.MkdirAll("docs/features", 0755) //nolint:errcheck
	os.WriteFile(".workflow/dep-edge-cases.md", []byte("- important dep lesson\n"), 0644) //nolint:errcheck
	os.WriteFile("docs/features/beta.md", []byte("## Problem\ntext\n"), 0644)             //nolint:errcheck

	cfg := memCfg()
	memory.Capture("dep", "tests", cfg)

	out := planadvisor.Directive("beta", cfg)
	if out == "" {
		t.Fatal("expected non-empty directive from plan advisor")
	}
}

// SC-13: concurrent captures for different features → distinct files.
func TestMemoryIntegration_ConcurrentCapture(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0755)                                                       //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- alpha lesson unique\n"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/bravo-edge-cases.md", []byte("- bravo lesson unique\n"), 0644) //nolint:errcheck

	cfg := memCfg()
	done := make(chan struct{}, 2)
	go func() {
		memory.Capture("alpha", "tests", cfg)
		done <- struct{}{}
	}()
	go func() {
		memory.Capture("bravo", "tests", cfg)
		done <- struct{}{}
	}()
	<-done
	<-done

	if countMdFiles(t) != 2 {
		t.Fatalf("expected 2 distinct entry files after concurrent capture (SC-13), got %d", countMdFiles(t))
	}
}
