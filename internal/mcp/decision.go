package mcp

import "github.com/samuelnp/centinela/internal/verdict"

// Decision tiers, worst-first. A harness maps block onto its pre-write deny.
const (
	Block = "block"
	Warn  = "warn"
	Allow = "allow"
)

// DecideGates is the gates-scope advisory decision.
func DecideGates(p *verdict.Packet) string {
	if p == nil {
		return Allow
	}
	if p.Summary.Gates.Fail > 0 {
		return Block
	}
	if p.Summary.Gates.Warn > 0 {
		return Warn
	}
	return Allow
}

// DecideVerify is the verify-scope advisory decision.
func DecideVerify(p *verdict.Packet) string {
	if p == nil {
		return Allow
	}
	if p.Summary.Verify.Fail > 0 {
		return Block
	}
	if p.Summary.Verify.Warn > 0 {
		return Warn
	}
	return Allow
}

// Combine reduces scope decisions to the worst (block > warn > allow). This is
// what the shim runs over the per-tool decisions; it equals Decide on the same
// packet, which is the parity guarantee.
func Combine(decisions ...string) string {
	worst := Allow
	for _, d := range decisions {
		if d == Block {
			return Block
		}
		if d == Warn {
			worst = Warn
		}
	}
	return worst
}

// Decide is the overall verdict for a packet (gates + verify combined).
func Decide(p *verdict.Packet) string {
	return Combine(DecideGates(p), DecideVerify(p))
}
