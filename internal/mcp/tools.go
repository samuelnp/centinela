package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// handleRules answers read_rules with the governing rule surface.
func (d Deps) handleRules(_ context.Context, _ *sdk.CallToolRequest, _ RulesInput) (*sdk.CallToolResult, RulesOutput, error) {
	out := d.Rules()
	out.Schema = SchemaVersion
	out.Gates = nz(out.Gates)
	out.Locales = nz(out.Locales)
	return nil, out, nil
}

// handleGates answers run_gates with gate lines and the gates-scope decision.
func (d Deps) handleGates(_ context.Context, _ *sdk.CallToolRequest, in FeatureInput) (*sdk.CallToolResult, GatesOutput, error) {
	p, err := d.Verdict(in.Feature)
	if err != nil {
		return nil, GatesOutput{}, err
	}
	return nil, GatesOutput{Schema: SchemaVersion, Decision: DecideGates(p), Gates: nz(p.Gates)}, nil
}

// handleVerify answers verify_claims with check lines and the verify-scope decision.
func (d Deps) handleVerify(_ context.Context, _ *sdk.CallToolRequest, in VerifyInput) (*sdk.CallToolResult, VerifyOutput, error) {
	p, err := d.Verdict(in.Feature)
	if err != nil {
		return nil, VerifyOutput{}, err
	}
	return nil, VerifyOutput{Schema: SchemaVersion, Decision: DecideVerify(p), Checks: nz(p.Verify)}, nil
}

// handleState answers workflow_state with run provenance and the evidence index.
func (d Deps) handleState(_ context.Context, _ *sdk.CallToolRequest, in FeatureInput) (*sdk.CallToolResult, StateOutput, error) {
	p, err := d.Verdict(in.Feature)
	if err != nil {
		return nil, StateOutput{}, err
	}
	return nil, StateOutput{Schema: SchemaVersion, Run: p.Run, Evidence: nz(p.Evidence)}, nil
}
