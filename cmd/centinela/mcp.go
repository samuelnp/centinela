package main

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	mcpgov "github.com/samuelnp/centinela/internal/mcp"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol governance server (advisory verdict over stdio)",
}

var mcpServeCmd = &cobra.Command{
	Use:           "serve",
	Short:         "Run the Centinela MCP governance server on stdio (centinela.mcp/v1)",
	RunE:          runMcpServe,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}

func runMcpServe(_ *cobra.Command, _ []string) error {
	srv := mcpgov.NewServer(mcpgov.Deps{Verdict: mcpVerdict, Rules: mcpRules})
	return srv.Run(context.Background(), &sdk.StdioTransport{})
}

// mcpVerdict assembles the verdict packet for a feature ("" = active feature),
// reusing the exact Deps wiring `centinela verdict` uses so MCP == native.
func mcpVerdict(feature string) (*verdict.Packet, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	wf := workflowForFeature(feature)
	if wf == nil {
		return nil, fmt.Errorf("no active feature; pass a feature slug")
	}
	deps := verdict.Deps{
		Gates: gates.RunAll,
		Verify: func(f, s string, c *config.Config) verify.VerificationResult {
			return verify.Verify(f, s, c, verify.Deps{Root: verifyRoot(), Runner: verify.NewExecRunner()})
		},
		Evidence: verdict.EvidenceIndex,
		Now:      time.Now().UTC().Format(time.RFC3339),
	}
	return verdict.AssembleVerdict(wf.Feature, cfg, wf, deps), nil
}

func workflowForFeature(feature string) *workflow.Workflow {
	if feature == "" {
		return activeWorkflow(mustGetwd())
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return nil
	}
	return wf
}
