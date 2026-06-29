package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
)

// exitMcpShim is overridable in tests; production uses os.Exit so a block
// verdict maps onto the harness pre-write deny (exit 2), like the native hook.
var exitMcpShim = os.Exit

var mcpShimCmd = &cobra.Command{
	Use:           "shim [feature]",
	Short:         "Obtain a verdict via the MCP server; exit 2 (deny) on block, 0 otherwise",
	RunE:          runMcpShim,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	mcpCmd.AddCommand(mcpShimCmd)
}

func runMcpShim(_ *cobra.Command, args []string) error {
	feature := ""
	if len(args) > 0 {
		feature = args[0]
	}
	ctx := context.Background()
	sess, err := mcpConnectSelf(ctx)
	if err != nil {
		return err
	}
	defer sess.Close() //nolint:errcheck
	gd, err := shimDecision(ctx, sess, "run_gates", feature)
	if err != nil {
		return err
	}
	vd, err := shimDecision(ctx, sess, "verify_claims", feature)
	if err != nil {
		return err
	}
	decision := mcpgov.Combine(gd, vd)
	if decision == mcpgov.Block {
		fmt.Fprintln(os.Stderr, "centinela mcp: block — write denied")
		exitMcpShim(2)
		return nil
	}
	fmt.Println("centinela mcp: " + decision)
	return nil
}

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
