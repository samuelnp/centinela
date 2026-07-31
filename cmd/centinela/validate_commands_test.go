package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/acceptance"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

const acceptCmd = "npx cucumber-js"
const skipOut = "3 scenarios (1 skipped, 2 passed)\n"

func cfgWithPolicy(policy string) *config.Config {
	c := &config.Config{}
	c.Validate.AcceptanceSkipPolicy = policy
	return c
}

// AC2 + the policy matrix, all on an acceptance-classified command that exits 0.
func TestCommandVerdict_AcceptanceSkipPolicy(t *testing.T) {
	cases := []struct {
		policy string
		want   gates.Status
	}{
		{config.AcceptanceSkipFail, gates.Fail},
		{"", gates.Fail},
		{config.AcceptanceSkipWarn, gates.Warn},
		{config.AcceptanceSkipOff, gates.Pass},
	}
	for _, c := range cases {
		t.Run("policy="+c.policy, func(t *testing.T) {
			status, detail := commandVerdict(cfgWithPolicy(c.policy), acceptCmd, true, skipOut)
			if status != c.want {
				t.Fatalf("status = %v, want %v (detail %q)", status, c.want, detail)
			}
			if c.want == gates.Pass {
				return
			}
			if !strings.Contains(detail, "1 skipped") || !strings.Contains(detail, acceptCmd) {
				t.Fatalf("message must name the counts and the command, got %q", detail)
			}
		})
	}
}

// E6: the exit code wins. No skip analysis runs, and the message does not claim
// the failure was caused by skips.
func TestCommandVerdict_NonZeroExitWinsOverSkips(t *testing.T) {
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail), acceptCmd, false, skipOut)
	if status != gates.Fail {
		t.Fatalf("a non-zero exit must fail, got %v", status)
	}
	if strings.Contains(detail, "skipped") {
		t.Fatalf("the exit failure must not be re-labelled as a skip verdict: %q", detail)
	}
}

// AC5: a NON-acceptance command reporting skips passes untouched.
func TestCommandVerdict_NonAcceptanceCommandWithSkipsPasses(t *testing.T) {
	out := "=== RUN   TestA\n--- SKIP: TestA (0.00s)\nPASS\nok  \tp\t0.1s\n"
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail), "go test ./internal/...", true, out)
	if status != gates.Pass || detail != "" {
		t.Fatalf("a non-acceptance command must be untouched, got %v / %q", status, detail)
	}
}

// AC3: an unparseable acceptance report warns and does not fail the run.
func TestCommandVerdict_UnparseableWarnsWithoutFailing(t *testing.T) {
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail), acceptCmd, true, "Ran 12 examples\n")
	if status != gates.Warn {
		t.Fatalf("unparseable output must warn, got %v", status)
	}
	if !strings.Contains(detail, "could not be parsed") || !strings.Contains(detail, "-json") {
		t.Fatalf("warning must name the limitation and a remedy, got %q", detail)
	}
}

// The config and the stdlib-only acceptance leaf restate the same three
// literals; this pins them equal so they cannot drift apart.
func TestSkipPolicyConstantsAgreeAcrossPackages(t *testing.T) {
	pairs := [][2]string{
		{config.AcceptanceSkipFail, acceptance.PolicyFail},
		{config.AcceptanceSkipWarn, acceptance.PolicyWarn},
		{config.AcceptanceSkipOff, acceptance.PolicyOff},
	}
	for _, p := range pairs {
		if p[0] != p[1] {
			t.Fatalf("policy constant drift: config %q vs acceptance %q", p[0], p[1])
		}
	}
}
