package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

var hookPlanAdvisorCmd = &cobra.Command{
	Use:   "plan-advisor",
	Short: "Hook: adaptive planning advisor during the plan step",
	RunE:  runHookPlanAdvisor,
}

func init() {
	hookCmd.AddCommand(hookPlanAdvisorCmd)
}

func runHookPlanAdvisor(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck
	cfg, err := config.LoadForProfile()
	if err != nil {
		// Hooks must never break the host session: warn and keep going with the
		// strict-pinned stand-in LoadForProfile returned.
		fmt.Println("config warning: " + err.Error())
	}
	if config.IsHeadless(cfg) {
		return nil
	}
	for _, wf := range loadActiveWorkflows() {
		if wf.CurrentStep != "plan" {
			continue
		}
		if out := planadvisor.Directive(wf.Feature, cfg); out != "" {
			fmt.Println(out)
		}
	}
	return nil
}
