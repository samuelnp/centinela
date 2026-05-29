package verify

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/evidence"
)

const covOut = "ok  pkg/a  coverage: 90.0% of statements\nok  pkg/b  coverage: 80.0% of statements\n"

func TestCheckCoverage(t *testing.T) {
	cases := []struct {
		name     string
		coverage *float64
		out      RunOutcome
		want     Status
	}{
		{"within-tolerance-pass", cov(85.0), RunOutcome{Output: covOut}, StatusPass},
		{"overclaim-fail", cov(92.0), RunOutcome{Output: covOut}, StatusFail},
		{"absent-skip", nil, RunOutcome{Output: covOut}, StatusSkip},
		{"timeout", cov(85.0), RunOutcome{TimedOut: true}, StatusTimeout},
		{"start-err", cov(85.0), RunOutcome{StartErr: errors.New("boom")}, StatusConfigError},
		{"no-figures", cov(85.0), RunOutcome{Output: "no coverage here"}, StatusConfigError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &evidence.RoleEvidence{Coverage: tc.coverage}
			deps := Deps{Runner: &fakeRunner{def: tc.out}}
			got := checkCoverage(cfgWithCmds("go test"), deps, "qa", ev, time.Second)
			if got.Status != tc.want {
				t.Fatalf("status = %q want %q (detail %q)", got.Status, tc.want, got.Detail)
			}
		})
	}
}

func TestCheckCoverageWithinTolerance(t *testing.T) {
	// claimed 85.0, measured mean of (84.95, 84.95) = 84.95; gap 0.05% < 0.1%.
	out := RunOutcome{Output: "coverage: 84.95% of statements\ncoverage: 84.95% of statements\n"}
	got := checkCoverage(cfgWithCmds("go test"), Deps{Runner: &fakeRunner{def: out}}, "qa", &evidence.RoleEvidence{Coverage: cov(85.0)}, time.Second)
	if got.Status != StatusPass {
		t.Fatalf("near-tolerance should pass, got %q / %q", got.Status, got.Detail)
	}
}

func TestMeanCoverage(t *testing.T) {
	mean, ok := meanCoverage(covOut)
	if !ok || mean != 85.0 {
		t.Fatalf("mean = %v ok=%v want 85.0", mean, ok)
	}
	if _, ok := meanCoverage(""); ok {
		t.Fatal("empty output should report ok=false")
	}
	// A figure that overflows float64 parsing reports ok=false.
	huge := "coverage: " + strings.Repeat("9", 400) + "e999999% of statements"
	if _, ok := meanCoverage(huge); ok {
		t.Fatal("unparseable figure should report ok=false")
	}
}
