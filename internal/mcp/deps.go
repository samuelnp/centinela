package mcp

import "github.com/samuelnp/centinela/internal/verdict"

// Deps injects the governance engines so the tool handlers stay pure and
// testable. The cmd layer wires the real implementations: Verdict wraps
// verdict.AssembleVerdict (the same path `centinela verdict` uses, so MCP and
// native verdicts share one assembler), and Rules reads the config rule surface.
type Deps struct {
	// Verdict assembles the full packet for a feature ("" = active feature).
	Verdict func(feature string) (*verdict.Packet, error)
	// Rules returns the governing rule surface (schema is stamped by the handler).
	Rules func() RulesOutput
}
