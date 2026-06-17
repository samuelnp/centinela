package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/autostart"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

var hookAutostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Hook: auto-start new feature workflow from prompt intent",
	RunE:  runHookAutostart,
}

func init() {
	hookCmd.AddCommand(hookAutostartCmd)
}

func runHookAutostart(_ *cobra.Command, _ []string) error {
	raw, _ := io.ReadAll(os.Stdin)
	if len(loadActiveWorkflows()) > 0 || !autostart.ShouldStart(autostart.ExtractPrompt(raw)) {
		return nil
	}
	feature := uniqueFeatureName(autostart.DeriveFeature(autostart.ExtractPrompt(raw)))
	order, err := workflowOrderForFeature(feature)
	if err != nil {
		return nil
	}
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	cfg, _ := config.Load()
	// No flags in the hook path: env/config still resolve inside ResolveStart.
	decision := workflow.ResolveStart("", "", cfg)
	wf := workflow.NewWithOrder(feature, order, decision.EffectiveProfile)
	wf.EnforcementProfile = decision.PinnedProfile
	wf.DriverModel = decision.DriverModel
	if err := workflow.Save(wf); err != nil {
		return nil
	}
	fmt.Printf("CENTINELA DIRECTIVE: auto-started workflow %q from prompt intent.\n", feature)
	return nil
}

func uniqueFeatureName(base string) string {
	name := base
	for i := 2; ; i++ {
		if _, err := os.Stat(workflow.FilePath(name)); os.IsNotExist(err) {
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
	}
}
