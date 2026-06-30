package integration_test

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/verdict"
)

// Integration: a real MCP client/server pair (in-memory transport) exercises the
// run_gates handler end-to-end; a fail-laden packet yields a block decision.
func TestMcpServerRunGatesInMemory(t *testing.T) {
	block := &verdict.Packet{}
	block.Summary.Gates = verdict.Counts{Fail: 1}
	deps := mcpgov.Deps{
		Verdict: func(string) (*verdict.Packet, error) { return block, nil },
		Rules:   func() mcpgov.RulesOutput { return mcpgov.RulesOutput{} },
	}

	ctx := context.Background()
	clientT, serverT := sdk.NewInMemoryTransports()
	server := mcpgov.NewServer(deps)
	go func() { _ = server.Run(ctx, serverT) }()

	client := sdk.NewClient(&sdk.Implementation{Name: "it", Version: "v1"}, nil)
	sess, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close() //nolint:errcheck

	res, err := sess.CallTool(ctx, &sdk.CallToolParams{
		Name: "run_gates", Arguments: map[string]any{"feature": "demo"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Schema   string `json:"schema"`
		Decision string `json:"decision"`
	}
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
				t.Fatal(err)
			}
		}
	}
	if out.Schema != "centinela.mcp/v1" || out.Decision != mcpgov.Block {
		t.Fatalf("unexpected tool result: %+v", out)
	}
}
