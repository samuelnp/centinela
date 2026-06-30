package main

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	mcpgov "github.com/samuelnp/centinela/internal/mcp"
)

// exitMcpShim is overridable in tests; production uses os.Exit so a block
// verdict maps onto the harness pre-write deny (exit 2), like the native hook.
var exitMcpShim = os.Exit

// mcpConnect is the connection seam — overridable in tests with an in-memory
// session so the decide/deny path is exercised without spawning a subprocess.
var mcpConnect = mcpConnectSelf

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
	sess, err := mcpConnect(ctx)
	if err != nil {
		return err
	}
	defer sess.Close() //nolint:errcheck
	decision, err := shimDecide(ctx, sess, feature)
	if err != nil {
		return err
	}
	if decision == mcpgov.Block {
		fmt.Fprintln(os.Stderr, "centinela mcp: block — write denied")
		exitMcpShim(2)
		return nil
	}
	fmt.Println("centinela mcp: " + decision)
	return nil
}

// shimDecide calls run_gates + verify_claims and combines their decisions into
// the overall verdict (worst wins), equalling Decide on the same packet.
func shimDecide(ctx context.Context, sess *sdk.ClientSession, feature string) (string, error) {
	gd, err := shimDecision(ctx, sess, "run_gates", feature)
	if err != nil {
		return "", err
	}
	vd, err := shimDecision(ctx, sess, "verify_claims", feature)
	if err != nil {
		return "", err
	}
	return mcpgov.Combine(gd, vd), nil
}
