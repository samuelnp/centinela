package acceptance

import (
	"strings"
	"testing"
)

// wholeRepoVerbose is the exact shape `go test -v ./...` produces: one
// contiguous, terminator-delimited block per package. The unit tier skips
// (docker / -short / build constraint), the acceptance tier does not.
const wholeRepoVerbose = `=== RUN   TestAcceptOK
--- PASS: TestAcceptOK (0.00s)
PASS
ok  	gv/tests/acceptance	2.03s
=== RUN   TestUnitSkip
    u_test.go:5: requires docker
--- SKIP: TestUnitSkip (0.00s)
=== RUN   TestUnitOK
--- PASS: TestUnitOK (0.00s)
PASS
ok  	gv/unitpkg	1.00s
`

// FINDING 1, the over-block direction. A unit-tier skip in a whole-repo command
// must never fail an acceptance gate — AC5 has to hold per TIER, not merely per
// command string.
func TestJudge_WholeRepoCommandIgnoresNonAcceptanceSkips(t *testing.T) {
	v, detail := Judge("go test -v ./...", wholeRepoVerbose, PolicyFail)
	if v != VerdictPass {
		t.Fatalf("a unit-tier skip must not fail a whole-repo command: %v (%s)", v, detail)
	}
}

// The same command, with the skip in the ACCEPTANCE tier, still fails: the fix
// narrows attribution, it does not disarm detection.
func TestJudge_WholeRepoCommandStillFailsOnAcceptanceSkips(t *testing.T) {
	out := strings.Replace(wholeRepoVerbose,
		"--- PASS: TestAcceptOK (0.00s)",
		"--- SKIP: TestAcceptOK (0.00s)", 1)
	v, detail := Judge("go test -v ./...", out, PolicyFail)
	if v != VerdictFail {
		t.Fatalf("an acceptance-tier skip must still fail: %v (%s)", v, detail)
	}
	if !strings.Contains(detail, "1 skipped") || !strings.Contains(detail, AcceptancePath) {
		t.Fatalf("detail must name the count and the attribution, got %q", detail)
	}
}

// Same split for `go test -json ./...`, where the package rides on every event.
func TestJudge_WholeRepoJSONAttributesByPackage(t *testing.T) {
	unit := `{"Action":"skip","Package":"gv/unitpkg","Test":"TestUnitSkip"}
{"Action":"pass","Package":"gv/tests/acceptance","Test":"TestA"}
`
	if v, d := Judge("go test -json ./...", unit, PolicyFail); v != VerdictPass {
		t.Fatalf("a unit-package skip must not fail a whole-repo run: %v (%s)", v, d)
	}
	accept := `{"Action":"skip","Package":"gv/tests/acceptance","Test":"TestA"}
{"Action":"pass","Package":"gv/unitpkg","Test":"TestUnitOK"}
`
	if v, d := Judge("go test -json ./...", accept, PolicyFail); v != VerdictFail {
		t.Fatalf("an acceptance-package skip must still fail: %v (%s)", v, d)
	}
}

// An acceptance-SCOPED command owns everything it ran, so no filtering applies.
func TestJudge_AcceptanceScopedCommandCountsEveryPackage(t *testing.T) {
	out := "--- SKIP: TestA (0.00s)\nPASS\nok  \tgv/features\t0.1s\n"
	if v, d := Judge("go test ./tests/acceptance/...", out, PolicyFail); v != VerdictFail {
		t.Fatalf("an acceptance-scoped command must count its own skips: %v (%s)", v, d)
	}
}

// A whole-repo run with no acceptance package must not be failed for "executed
// no scenarios" — that rule only means something when the command was scoped.
func TestJudge_WholeRepoWithNoAcceptancePackageIsNotZeroScenarios(t *testing.T) {
	out := "=== RUN   TestUnitOK\n--- PASS: TestUnitOK (0.00s)\nPASS\nok  \tgv/unitpkg\t0.1s\n"
	if v, d := Judge("go test -v ./...", out, PolicyFail); v != VerdictPass {
		t.Fatalf("no attributable acceptance work is not proof nothing ran: %v (%s)", v, d)
	}
}
