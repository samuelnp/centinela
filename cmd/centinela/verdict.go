package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

var verdictCmd = &cobra.Command{
	Use:           "verdict <feature>",
	Short:         "Emit a deterministic machine-readable JSON verdict for a feature",
	Args:          cobra.ExactArgs(1),
	RunE:          runVerdict,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(verdictCmd)
}

func runVerdict(_ *cobra.Command, args []string) error {
	feature := args[0]
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}
	deps := verdict.Deps{
		Gates: gates.RunAll,
		Verify: func(f, s string, c *config.Config) verify.VerificationResult {
			return verify.Verify(f, s, c, verify.Deps{Root: verifyRoot(), Runner: verify.NewExecRunner()})
		},
		Evidence: verdict.EvidenceIndex,
		Now:      time.Now().UTC().Format(time.RFC3339),
	}
	pkt := verdict.AssembleVerdict(feature, cfg, wf, deps)
	data, err := json.MarshalIndent(pkt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, string(data))
	if pkt.Summary.ExitCode != 0 {
		return fmt.Errorf("verdict: fail (%d gate / %d verify failures)",
			pkt.Summary.Gates.Fail, pkt.Summary.Verify.Fail)
	}
	return nil
}
