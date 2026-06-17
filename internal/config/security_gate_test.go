package config

import (
	"strings"
	"testing"
)

// TestNormalizeSecurityGate_DisabledIsNoOp verifies that disabled config is
// returned unchanged (no tool defaults are applied).
func TestNormalizeSecurityGate_DisabledIsNoOp(t *testing.T) {
	in := SecurityGateConfig{Enabled: false}
	out := NormalizeSecurityGate(in)
	if out.Secrets.Tool != "" || len(out.Vuln.Tools) != 0 {
		t.Fatalf("disabled: unexpected fields set: %+v", out)
	}
}

// TestNormalizeSecurityGate_EnabledDefaultsSecretsTool verifies that an
// enabled config with no secrets tool gets gitleaks as the default.
func TestNormalizeSecurityGate_EnabledDefaultsSecretsTool(t *testing.T) {
	out := NormalizeSecurityGate(SecurityGateConfig{Enabled: true})
	if out.Secrets.Tool != SecretsTool {
		t.Fatalf("expected %q, got %q", SecretsTool, out.Secrets.Tool)
	}
}

// TestNormalizeSecurityGate_EnabledDefaultsVulnTools verifies that when no
// vuln tools are configured the full VulnTools set is applied.
func TestNormalizeSecurityGate_EnabledDefaultsVulnTools(t *testing.T) {
	out := NormalizeSecurityGate(SecurityGateConfig{Enabled: true})
	if len(out.Vuln.Tools) != len(VulnTools) {
		t.Fatalf("expected %v, got %v", VulnTools, out.Vuln.Tools)
	}
}

// TestNormalizeSecurityGate_TrimsWhitespace verifies leading/trailing
// whitespace is stripped from tool names and allowlist entries.
func TestNormalizeSecurityGate_TrimsWhitespace(t *testing.T) {
	in := SecurityGateConfig{
		Enabled: true,
		Secrets: SecretsConfig{Tool: "  gitleaks  ", Allowlist: []string{"  ", "ok  "}},
		Vuln:    VulnConfig{Tools: []string{"  govulncheck  "}},
	}
	out := NormalizeSecurityGate(in)
	if out.Secrets.Tool != "gitleaks" {
		t.Fatalf("tool not trimmed: %q", out.Secrets.Tool)
	}
	if len(out.Secrets.Allowlist) != 1 || out.Secrets.Allowlist[0] != "ok" {
		t.Fatalf("allowlist not trimmed: %v", out.Secrets.Allowlist)
	}
}

// TestValidateSecurityGate_DisabledIsNoOp confirms disabled config is always
// accepted regardless of tool names.
func TestValidateSecurityGate_DisabledIsNoOp(t *testing.T) {
	cfg := SecurityGateConfig{Enabled: false, Secrets: SecretsConfig{Tool: "unknown"}}
	if err := validateSecurityGate(cfg); err != nil {
		t.Fatalf("disabled: unexpected error: %v", err)
	}
}

// TestValidateSecurityGate_UnknownSecretsTool rejects an unknown secrets tool.
func TestValidateSecurityGate_UnknownSecretsTool(t *testing.T) {
	cfg := SecurityGateConfig{
		Enabled: true,
		Secrets: SecretsConfig{Tool: "trivy"},
		Vuln:    VulnConfig{Tools: VulnTools},
	}
	err := validateSecurityGate(cfg)
	if err == nil || !strings.Contains(err.Error(), "trivy") {
		t.Fatalf("expected rejection of unknown tool, got %v", err)
	}
}

// TestValidateSecurityGate_UnknownVulnTool rejects an unknown vuln tool.
func TestValidateSecurityGate_UnknownVulnTool(t *testing.T) {
	cfg := SecurityGateConfig{
		Enabled: true,
		Secrets: SecretsConfig{Tool: SecretsTool},
		Vuln:    VulnConfig{Tools: []string{"govulncheck", "snyk"}},
	}
	err := validateSecurityGate(cfg)
	if err == nil || !strings.Contains(err.Error(), "snyk") {
		t.Fatalf("expected rejection of snyk, got %v", err)
	}
}

// TestValidateSecurityGate_ValidConfigPasses confirms a fully valid enabled
// config passes without error.
func TestValidateSecurityGate_ValidConfigPasses(t *testing.T) {
	cfg := SecurityGateConfig{
		Enabled: true,
		Secrets: SecretsConfig{Tool: SecretsTool},
		Vuln:    VulnConfig{Tools: VulnTools},
	}
	if err := validateSecurityGate(cfg); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
}
