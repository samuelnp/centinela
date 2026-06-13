package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// Acceptance: specs/headless-governance.feature

func hgCfg(enabled, detectCI bool) *config.Config {
	c := &config.Config{}
	c.Headless = config.HeadlessConfig{Enabled: enabled, DetectCI: detectCI}
	return c
}

// Scenario: CI auto-detect opt-in makes the run headless
func TestHG_CIAutoDetectOptIn(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "true")
	if !config.IsHeadless(hgCfg(false, true)) {
		t.Fatal("detect_ci with CI=true must be headless")
	}
}

// Scenario: CI present but detect_ci off is not headless (back-compat)
func TestHG_CIPresentDetectOff(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "true")
	if config.IsHeadless(hgCfg(false, false)) {
		t.Fatal("CI present with detect_ci off must NOT be headless")
	}
}

// Scenario: Zero-config default is not headless
func TestHG_ZeroConfigDefaultOff(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "")
	if config.IsHeadless(hgCfg(false, false)) {
		t.Fatal("zero-config default must not be headless")
	}
}

// Scenario: Env override beats config off
func TestHG_EnvOverrideBeatsConfigOff(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	if !config.IsHeadless(hgCfg(false, false)) {
		t.Fatal("env=1 must win over config enabled=false")
	}
}

// Scenario: Empty env value falls through to config and detect_ci
func TestHG_EmptyEnvFallsThrough(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	t.Setenv("CI", "")
	if config.IsHeadless(hgCfg(false, false)) {
		t.Fatal("empty env must fall through to (off) config and detect_ci")
	}
}
