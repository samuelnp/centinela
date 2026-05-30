package integration_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

func buildGateCfg(enabled bool) *config.Config {
	return &config.Config{Gates: config.GatesConfig{
		Build: config.BuildGateConfig{
			Enabled: enabled,
			Command: "go version",
			Targets: []config.BuildTarget{{GOOS: "linux", GOARCH: "amd64"}},
		},
	}}
}

func hasBuildGate(results []gates.Result) bool {
	for _, r := range results {
		if r.Name == "G-Build: Cross-Compile" {
			return true
		}
	}
	return false
}

func TestRunWithFilter_IncludesBuildGateWhenEnabled(t *testing.T) {
	results := gates.RunWithFilter(buildGateCfg(true), nil)
	if !hasBuildGate(results) {
		t.Fatalf("expected G-Build result when enabled, got %+v", results)
	}
}

func TestRunWithFilter_ExcludesBuildGateWhenDisabled(t *testing.T) {
	results := gates.RunWithFilter(buildGateCfg(false), nil)
	if hasBuildGate(results) {
		t.Fatalf("expected no G-Build result when disabled, got %+v", results)
	}
}
