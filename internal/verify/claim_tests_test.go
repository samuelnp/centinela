package verify

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgWithCmds(cmds ...string) *config.Config {
	c := &config.Config{}
	c.Validate.Commands = cmds
	c.Verify.TimeoutSeconds = 60
	c.Verify.CoverageTolerance = 0.001
	return c
}

func TestCheckTestsPass(t *testing.T) {
	cases := []struct {
		name string
		cmds []string
		out  RunOutcome
		want Status
	}{
		{"pass", []string{"go test"}, RunOutcome{ExitCode: 0}, StatusPass},
		{"fail", []string{"go test"}, RunOutcome{ExitCode: 1}, StatusFail},
		{"timeout", []string{"go test"}, RunOutcome{TimedOut: true}, StatusTimeout},
		{"missing-binary", []string{"frobnicate"}, RunOutcome{StartErr: errors.New(`exec: "frobnicate": not found`)}, StatusConfigError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := Deps{Runner: &fakeRunner{def: tc.out}}
			got := checkTestsPass(cfgWithCmds(tc.cmds...), deps, "qa-senior", time.Second)
			if got.Status != tc.want {
				t.Fatalf("status = %q want %q (detail %q)", got.Status, tc.want, got.Detail)
			}
		})
	}
}

func TestCheckTestsPassNoCommands(t *testing.T) {
	got := checkTestsPass(cfgWithCmds(), Deps{Runner: &fakeRunner{}}, "qa", time.Second)
	if got.Status != StatusConfigError || !strings.Contains(got.Detail, "validate.commands") {
		t.Fatalf("expected config error naming validate.commands, got %q / %q", got.Status, got.Detail)
	}
}

func TestCheckTestsPassPriorRun(t *testing.T) {
	r := &fakeRunner{def: RunOutcome{ExitCode: 0}}
	deps := Deps{Runner: r, PriorTestRun: &RunOutcome{ExitCode: 1}}
	got := checkTestsPass(cfgWithCmds("go test"), deps, "qa", time.Second)
	if got.Status != StatusFail {
		t.Fatalf("prior run should be reused (fail), got %q", got.Status)
	}
	if len(r.calls) != 0 {
		t.Fatalf("runner should not be invoked when PriorTestRun set, calls=%v", r.calls)
	}
}

func TestMissingBinary(t *testing.T) {
	if got := missingBinary(errors.New("exec: bad: nope")); !strings.Contains(got, "nope") {
		t.Fatalf("missingBinary lost cause: %q", got)
	}
	if got := missingBinary(errors.New("flat")); got != "flat" {
		t.Fatalf("missingBinary = %q", got)
	}
}
