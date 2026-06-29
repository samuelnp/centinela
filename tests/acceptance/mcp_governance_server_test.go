package acceptance_test

// Acceptance: specs/mcp-governance-server.feature

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/verdict"
)

// Scenario: a zero-integration harness lists tools and reads a versioned verdict.
func TestAccMcpZeroIntegrationHarness(t *testing.T) {
	bin := buildMcpBin(t)
	sess := connectMcp(t, bin, mcpRepo(t, false))
	defer sess.Close() //nolint:errcheck
	lt, err := sess.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tool := range lt.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{"read_rules", "run_gates", "verify_claims", "workflow_state"} {
		if !got[want] {
			t.Errorf("missing tool %q", want)
		}
	}
	if !strings.Contains(toolText(t, sess, "read_rules", map[string]any{}), `"schema":"centinela.mcp/v1"`) {
		t.Error("read_rules missing versioned schema")
	}
}

func decision(t *testing.T, text string) string {
	t.Helper()
	var out struct {
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("decode decision: %v", err)
	}
	return out.Decision
}

// Scenario: the shim denies a write on block and allows on allow.
func TestAccMcpShimBlockAndAllow(t *testing.T) {
	bin := buildMcpBin(t)
	for _, tc := range []struct {
		name string
		want int
	}{{"block", 2}, {"allow", 0}} {
		c := exec.Command(bin, "mcp", "shim", "demo")
		c.Dir = mcpRepo(t, tc.name == "block")
		code := 0
		if err := c.Run(); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else {
				t.Fatalf("%s: run: %v", tc.name, err)
			}
		}
		if code != tc.want {
			t.Errorf("%s: shim exit=%d want %d", tc.name, code, tc.want)
		}
	}
}

// Scenario: the MCP verdict matches the native-hook verdict (parity).
func TestAccMcpParityWithNative(t *testing.T) {
	bin := buildMcpBin(t)
	dir := mcpRepo(t, true)
	sess := connectMcp(t, bin, dir)
	defer sess.Close() //nolint:errcheck
	mcpDecision := mcpgov.Combine(
		decision(t, toolText(t, sess, "run_gates", map[string]any{"feature": "demo"})),
		decision(t, toolText(t, sess, "verify_claims", map[string]any{"feature": "demo"})),
	)
	native := exec.Command(bin, "verdict", "demo")
	native.Dir = dir
	out, _ := native.Output()
	raw := out[strings.IndexByte(string(out), '{'):]
	var pkt verdict.Packet
	if err := json.Unmarshal([]byte(raw), &pkt); err != nil {
		t.Fatalf("parse native verdict: %v", err)
	}
	if got := mcpgov.Decide(&pkt); got != mcpDecision {
		t.Errorf("parity broken: mcp=%s native=%s", mcpDecision, got)
	}
}
