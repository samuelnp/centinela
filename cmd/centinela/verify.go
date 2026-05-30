package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <feature>",
	Short: "Independently verify a feature's evidence claims against ground truth",
	Args:  cobra.ExactArgs(1),
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(_ *cobra.Command, args []string) error {
	feature := args[0]
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}
	res := verify.Verify(feature, wf.CurrentStep, cfg, verify.Deps{
		Root:   verifyRoot(),
		Runner: verify.NewExecRunner(),
	})
	fmt.Println(ui.RenderVerification(res))
	if res.HasFailures() || res.HasWarnings() {
		return fmt.Errorf("verification did not pass cleanly for %q", feature)
	}
	return nil
}

// verifyRoot resolves the directory verification runs against: the active
// worktree root when invoked inside one, else the current directory.
func verifyRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	if _, root := worktree.DetectFeatureFromCwd(cwd); root != "" {
		return root
	}
	return cwd
}
