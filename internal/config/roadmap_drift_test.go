package config

import (
	"strings"
	"testing"
)

func TestNormalizeRoadmapDriftDefaultsToWarn(t *testing.T) {
	got := NormalizeRoadmapDrift(RoadmapDriftConfig{Enabled: true})
	if got.Severity != "warn" {
		t.Fatalf("unset severity must default to warn, got %q", got.Severity)
	}
}

func TestNormalizeRoadmapDriftTrims(t *testing.T) {
	got := NormalizeRoadmapDrift(RoadmapDriftConfig{Enabled: true, Severity: "  fail  "})
	if got.Severity != "fail" {
		t.Fatalf("severity should be trimmed, got %q", got.Severity)
	}
}

func TestValidateRoadmapDriftAccepts(t *testing.T) {
	for _, sev := range []string{"fail", "warn"} {
		if err := validateRoadmapDrift(RoadmapDriftConfig{Enabled: true, Severity: sev}); err != nil {
			t.Fatalf("severity %q must be accepted: %v", sev, err)
		}
	}
}

func TestValidateRoadmapDriftRejectsUnknown(t *testing.T) {
	err := validateRoadmapDrift(RoadmapDriftConfig{Enabled: true, Severity: "error"})
	if err == nil {
		t.Fatal("unknown severity must be rejected")
	}
	got := err.Error()
	if !strings.Contains(got, "severity") || !strings.Contains(got, "fail") || !strings.Contains(got, "warn") {
		t.Fatalf("error must name the field and valid values, got %q", got)
	}
}

func TestValidateRoadmapDriftNoopWhenDisabled(t *testing.T) {
	if err := validateRoadmapDrift(RoadmapDriftConfig{Enabled: false, Severity: "bogus"}); err != nil {
		t.Fatalf("bad severity must be a no-op when disabled, got %v", err)
	}
}
