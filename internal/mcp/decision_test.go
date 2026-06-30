package mcp

import (
	"testing"

	"github.com/samuelnp/centinela/internal/verdict"
)

func pkt(gFail, gWarn, vFail, vWarn int) *verdict.Packet {
	p := &verdict.Packet{}
	p.Summary.Gates = verdict.Counts{Fail: gFail, Warn: gWarn}
	p.Summary.Verify = verdict.Counts{Fail: vFail, Warn: vWarn}
	return p
}

func TestDecideScopes(t *testing.T) {
	cases := []struct {
		name             string
		p                *verdict.Packet
		gates, verify, c string
	}{
		{"clean", pkt(0, 0, 0, 0), Allow, Allow, Allow},
		{"gate warn", pkt(0, 1, 0, 0), Warn, Allow, Warn},
		{"gate fail", pkt(1, 0, 0, 0), Block, Allow, Block},
		{"verify fail", pkt(0, 0, 1, 0), Allow, Block, Block},
		{"verify warn", pkt(0, 0, 0, 2), Allow, Warn, Warn},
		{"both fail", pkt(1, 0, 1, 0), Block, Block, Block},
	}
	for _, tc := range cases {
		if got := DecideGates(tc.p); got != tc.gates {
			t.Errorf("%s: DecideGates=%s want %s", tc.name, got, tc.gates)
		}
		if got := DecideVerify(tc.p); got != tc.verify {
			t.Errorf("%s: DecideVerify=%s want %s", tc.name, got, tc.verify)
		}
		if got := Decide(tc.p); got != tc.c {
			t.Errorf("%s: Decide=%s want %s", tc.name, got, tc.c)
		}
	}
}

func TestDecideNilPacket(t *testing.T) {
	if DecideGates(nil) != Allow || DecideVerify(nil) != Allow {
		t.Fatal("nil packet must be allow")
	}
}

func TestCombineWorstWins(t *testing.T) {
	if Combine(Allow, Warn, Allow) != Warn {
		t.Fatal("warn should beat allow")
	}
	if Combine(Warn, Block) != Block {
		t.Fatal("block should beat warn")
	}
	if Combine() != Allow {
		t.Fatal("empty combine is allow")
	}
}

func TestNzCoalescesNil(t *testing.T) {
	var s []int
	if got := nz(s); got == nil || len(got) != 0 {
		t.Fatalf("nz(nil) should be empty non-nil, got %v", got)
	}
	if got := nz([]int{1}); len(got) != 1 {
		t.Fatalf("nz passthrough broken: %v", got)
	}
}
