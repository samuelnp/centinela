package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
)

// mcpConnectSelf launches this binary as `mcp serve` and connects an MCP client.
func mcpConnectSelf(ctx context.Context) (*sdk.ClientSession, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "centinela-shim", Version: mcpgov.SchemaVersion}, nil)
	return client.Connect(ctx, &sdk.CommandTransport{Command: exec.Command(self, "mcp", "serve")}, nil)
}

// shimDecision calls a tool and reads its "decision" field from the JSON result.
func shimDecision(ctx context.Context, sess *sdk.ClientSession, tool, feature string) (string, error) {
	res, err := sess.CallTool(ctx, &sdk.CallToolParams{Name: tool, Arguments: map[string]any{"feature": feature}})
	if err != nil {
		return "", err
	}
	if res.IsError {
		return "", fmt.Errorf("%s: tool error", tool)
	}
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			var out struct {
				Decision string `json:"decision"`
			}
			if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
				return "", err
			}
			return out.Decision, nil
		}
	}
	return "", fmt.Errorf("%s: no text content", tool)
}
