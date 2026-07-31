package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// wholeRepoVerboseOut is what `go test -v ./...` prints: one terminator-
// delimited block per package. Only the unit tier skips.
const wholeRepoVerboseOut = "=== RUN   TestAcceptOK\n--- PASS: TestAcceptOK (0.00s)\n" +
	"PASS\nok  \tgv/tests/acceptance\t2.03s\n" +
	"=== RUN   TestUnitSkip\n    u_test.go:5: requires docker\n" +
	"--- SKIP: TestUnitSkip (0.00s)\nPASS\nok  \tgv/unitpkg\t1.00s\n"

// FINDING 1 through the runner: following this tool's own "add -v" advice must
// not turn a unit-tier docker/-short/build-tag skip into a validate failure.
func TestCommandVerdict_WholeRepoVerboseUnitSkipStaysGreen(t *testing.T) {
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail),
		"go test -v ./...", true, wholeRepoVerboseOut)
	if status != gates.Pass {
		t.Fatalf("a unit-tier skip must not fail a whole-repo command, got %v (%q)", status, detail)
	}
}

// The mirror: an acceptance-tier skip in the same command still fails.
func TestCommandVerdict_WholeRepoVerboseAcceptanceSkipStillFails(t *testing.T) {
	out := strings.Replace(wholeRepoVerboseOut,
		"--- PASS: TestAcceptOK (0.00s)", "--- SKIP: TestAcceptOK (0.00s)", 1)
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail),
		"go test -v ./...", true, out)
	if status != gates.Fail {
		t.Fatalf("an acceptance-tier skip must still fail, got %v (%q)", status, detail)
	}
	if !strings.Contains(detail, "1 skipped") {
		t.Fatalf("detail must name the count, got %q", detail)
	}
}

// FINDING 2 through the runner: a clean cucumber summary must not hide a
// go-level skip printed in the same output.
func TestCommandVerdict_CucumberSummaryDoesNotHideGoSkips(t *testing.T) {
	out := "3 scenarios (3 passed)\n--- SKIP: TestGoLevelHidden (0.00s)\n" +
		"PASS\nok  \tx/tests/acceptance\t0.42s\n"
	status, detail := commandVerdict(cfgWithPolicy(config.AcceptanceSkipFail),
		"./run.sh # tests/acceptance", true, out)
	if status != gates.Fail {
		t.Fatalf("a go skip beside a clean summary must fail, got %v (%q)", status, detail)
	}
	if !strings.Contains(detail, "1 skipped") {
		t.Fatalf("detail must name the hidden skip, got %q", detail)
	}
}

// End-to-end through runValidateCommands, both directions in one place.
func TestRunValidateCommands_TierAttributionEndToEnd(t *testing.T) {
	cases := []struct {
		name, script string
		wantPass     bool
	}{
		{"unit-skip", `printf '%s' "--- SKIP: TestUnitSkip (0.00s)
ok  	gv/unitpkg	0.1s
" # go test ./...`, true},
		{"acceptance-skip", `printf '%s' "--- SKIP: TestAcceptSkip (0.00s)
ok  	gv/tests/acceptance	0.1s
" # go test ./...`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Validate.Commands = []string{c.script}
			cfg.Validate.AcceptanceSkipPolicy = config.AcceptanceSkipFail
			var passed bool
			out := captureStdout(t, func() { passed = runValidateCommands(cfg) })
			if passed != c.wantPass {
				t.Fatalf("passed = %v, want %v\n%s", passed, c.wantPass, out)
			}
		})
	}
}
