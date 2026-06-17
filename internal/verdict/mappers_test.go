package verdict

import (
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
)

// Every gate Status lowercases into the packet vocabulary.
func TestGateStatus_Lowercases(t *testing.T) {
	cases := map[gates.Status]string{
		gates.Pass:       "pass",
		gates.Fail:       "fail",
		gates.Warn:       "warn",
		gates.Skip:       "skip",
		gates.Status(99): "unknown",
	}
	for in, want := range cases {
		if got := gateStatus(in); got != want {
			t.Fatalf("gateStatus(%v) = %q, want %q", in, got, want)
		}
	}
}

// gateLine carries name/message/details and the lowercased status.
func TestGateLine_Maps(t *testing.T) {
	gl := gateLine(gates.Result{Name: "G1", Status: gates.Fail, Message: "m", Details: []string{"d"}})
	if gl.Name != "G1" || gl.Status != "fail" || gl.Message != "m" || len(gl.Details) != 1 {
		t.Fatalf("gateLine mapped wrong: %+v", gl)
	}
}

// Verify check statuses keep their native UPPERCASE vocabulary.
func TestCheckLine_PreservesUppercase(t *testing.T) {
	cl := checkLine(verify.Check{Role: "qa-senior", Claim: "tests-pass", Status: verify.StatusPass, Detail: "x"})
	if cl.Status != "PASS" || cl.Role != "qa-senior" || cl.Claim != "tests-pass" || cl.Detail != "x" {
		t.Fatalf("checkLine mapped wrong: %+v", cl)
	}
	if checkLine(verify.Check{Status: verify.StatusConfigError}).Status != "CONFIG-ERROR" {
		t.Fatal("CONFIG-ERROR must be preserved verbatim")
	}
}

// gateCounts tallies every status bucket.
func TestGateCounts(t *testing.T) {
	c := gateCounts([]gates.Result{
		{Status: gates.Pass}, {Status: gates.Pass}, {Status: gates.Fail}, {Status: gates.Warn}, {Status: gates.Skip},
	})
	if c.Pass != 2 || c.Fail != 1 || c.Warn != 1 || c.Skip != 1 {
		t.Fatalf("gateCounts = %+v", c)
	}
}

// verifyCounts maps the native Tally (pass, fail, skip, warn) into Counts.
func TestVerifyCounts(t *testing.T) {
	c := verifyCounts(vr(
		verify.Check{Status: verify.StatusPass},
		verify.Check{Status: verify.StatusFail},
		verify.Check{Status: verify.StatusSkip},
		verify.Check{Status: verify.StatusWarn},
	))
	if c.Pass != 1 || c.Fail != 1 || c.Skip != 1 || c.Warn != 1 {
		t.Fatalf("verifyCounts = %+v", c)
	}
}
