package mcp

import sdk "github.com/modelcontextprotocol/go-sdk/mcp"

// NewServer builds the Centinela MCP server with the four versioned governance
// tools. The server is advisory: every tool only reads (gates, claims, workflow
// state); none mutate or block.
func NewServer(d Deps) *sdk.Server {
	s := sdk.NewServer(&sdk.Implementation{Name: "centinela", Version: SchemaVersion}, nil)
	sdk.AddTool(s, &sdk.Tool{
		Name:        "read_rules",
		Description: "Return the governing rule surface (profile, archetype, file-size limit, enabled gates, locales).",
	}, d.handleRules)
	sdk.AddTool(s, &sdk.Tool{
		Name:        "run_gates",
		Description: "Run the gate suite and return gate results plus the gates-scope decision (allow/warn/block).",
	}, d.handleGates)
	sdk.AddTool(s, &sdk.Tool{
		Name:        "verify_claims",
		Description: "Re-derive the feature's evidence claims and return check results plus the verify-scope decision.",
	}, d.handleVerify)
	sdk.AddTool(s, &sdk.Tool{
		Name:        "workflow_state",
		Description: "Return the active feature's run provenance and on-disk role evidence index.",
	}, d.handleState)
	return s
}
