package config

import (
	"strings"
	"testing"
)

func TestNormalizeSpecTraceability_FillsDefaults(t *testing.T) {
	got := NormalizeSpecTraceability(SpecTraceabilityConfig{Enabled: true})
	if got.SpecDir != "specs" || got.TestDir != "tests/acceptance" || got.Severity != "fail" {
		t.Fatalf("defaults not applied: %+v", got)
	}
}

func TestNormalizeSpecTraceability_TrimsConfiguredValues(t *testing.T) {
	got := NormalizeSpecTraceability(SpecTraceabilityConfig{
		SpecDir: "  feats  ", TestDir: "  t/acc ", Severity: " warn ",
	})
	if got.SpecDir != "feats" || got.TestDir != "t/acc" || got.Severity != "warn" {
		t.Fatalf("trim wrong: %+v", got)
	}
}

func TestValidateSpecTraceability_RejectsUnknownSeverity(t *testing.T) {
	err := validateSpecTraceability(SpecTraceabilityConfig{Enabled: true, Severity: "loud"})
	if err == nil || !strings.Contains(err.Error(), "severity") {
		t.Fatalf("want severity error, got %v", err)
	}
}

func TestValidateSpecTraceability_AcceptsWarnAndFail(t *testing.T) {
	if validateSpecTraceability(SpecTraceabilityConfig{Enabled: true, Severity: "warn"}) != nil {
		t.Fatal("warn must be accepted")
	}
	if validateSpecTraceability(SpecTraceabilityConfig{Enabled: true, Severity: "fail"}) != nil {
		t.Fatal("fail must be accepted")
	}
}

func TestValidateSpecTraceability_DisabledIsNoOp(t *testing.T) {
	if validateSpecTraceability(SpecTraceabilityConfig{Enabled: false, Severity: "loud"}) != nil {
		t.Fatal("disabled gate must skip severity validation")
	}
}
