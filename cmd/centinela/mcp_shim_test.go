package main

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/verdict"
)

// inMemShimSession starts an in-memory MCP server with the given deps and returns
// a connected client session, so the shim's decide/deny path runs without a
// subprocess.
func inMemShimSession(t *testing.T, p *verdict.Packet) *sdk.ClientSession {
	t.Helper()
	deps := mcpgov.Deps{
		Verdict: func(string) (*verdict.Packet, error) { return p, nil },
		Rules:   func() mcpgov.RulesOutput { return mcpgov.RulesOutput{} },
	}
	ct, st := sdk.NewInMemoryTransports()
	go func() { _ = mcpgov.NewServer(deps).Run(context.Background(), st) }()
	client := sdk.NewClient(&sdk.Implementation{Name: "shimtest", Version: "v1"}, nil)
	sess, err := client.Connect(context.Background(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func withShimSession(t *testing.T, p *verdict.Packet) *int {
	t.Helper()
	sess := inMemShimSession(t, p)
	oldConn := mcpConnect
	mcpConnect = func(context.Context) (*sdk.ClientSession, error) { return sess, nil }
	code := -1
	oldExit := exitMcpShim
	exitMcpShim = func(c int) { code = c }
	t.Cleanup(func() { mcpConnect = oldConn; exitMcpShim = oldExit })
	return &code
}

func TestRunMcpShimBlockExitsTwo(t *testing.T) {
	block := &verdict.Packet{}
	block.Summary.Gates = verdict.Counts{Fail: 1}
	code := withShimSession(t, block)
	if err := runMcpShim(nil, []string{"demo"}); err != nil {
		t.Fatal(err)
	}
	if *code != 2 {
		t.Fatalf("block should exit 2, got %d", *code)
	}
}

func TestRunMcpShimAllowDoesNotExit(t *testing.T) {
	code := withShimSession(t, &verdict.Packet{}) // clean packet → allow
	if err := runMcpShim(nil, nil); err != nil {
		t.Fatal(err)
	}
	if *code != -1 {
		t.Fatalf("allow must not call exit, got code %d", *code)
	}
}

func TestRunMcpShimToolErrorPropagates(t *testing.T) {
	deps := mcpgov.Deps{
		Verdict: func(string) (*verdict.Packet, error) { return nil, errShim },
		Rules:   func() mcpgov.RulesOutput { return mcpgov.RulesOutput{} },
	}
	ct, st := sdk.NewInMemoryTransports()
	go func() { _ = mcpgov.NewServer(deps).Run(context.Background(), st) }()
	client := sdk.NewClient(&sdk.Implementation{Name: "e", Version: "v1"}, nil)
	sess, err := client.Connect(context.Background(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldConn := mcpConnect
	mcpConnect = func(context.Context) (*sdk.ClientSession, error) { return sess, nil }
	t.Cleanup(func() { mcpConnect = oldConn })
	if err := runMcpShim(nil, []string{"demo"}); err == nil {
		t.Fatal("expected the tool error to propagate")
	}
}

var errShim = fmt.Errorf("verdict boom")
