package unit_test

import (
	"testing"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/verdict"
)

// Unit: the advisory decision maps gate/verify outcomes to allow/warn/block,
// worst-wins, which is the contract the harness shim relies on.
func TestMcpDecisionMapping(t *testing.T) {
	mk := func(gf, gw, vf, vw int) *verdict.Packet {
		p := &verdict.Packet{}
		p.Summary.Gates = verdict.Counts{Fail: gf, Warn: gw}
		p.Summary.Verify = verdict.Counts{Fail: vf, Warn: vw}
		return p
	}
	cases := []struct {
		p    *verdict.Packet
		want string
	}{
		{mk(0, 0, 0, 0), mcpgov.Allow},
		{mk(0, 1, 0, 0), mcpgov.Warn},
		{mk(1, 0, 0, 0), mcpgov.Block},
		{mk(0, 0, 0, 3), mcpgov.Warn},
		{mk(0, 0, 2, 0), mcpgov.Block},
	}
	for i, c := range cases {
		if got := mcpgov.Decide(c.p); got != c.want {
			t.Errorf("case %d: Decide=%s want %s", i, got, c.want)
		}
	}
}
