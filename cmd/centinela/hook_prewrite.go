package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
)

var hookPrewriteCmd = &cobra.Command{
	Use:   "prewrite",
	Short: "Hook: block writes in the wrong workflow step",
	RunE:  runHookPrewrite,
}

func init() {
	hookCmd.AddCommand(hookPrewriteCmd)
}

type prewriteInput struct {
	ToolInput struct {
		FilePath  string `json:"file_path"`
		FilePath2 string `json:"filePath"`
	} `json:"tool_input"`
}

var exitPrewrite = os.Exit
var evalPrewrite = hookpolicy.EvaluatePrewrite

func runHookPrewrite(_ *cobra.Command, _ []string) error {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var input prewriteInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil
	}
	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolInput.FilePath2
	}
	if filePath == "" {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		// Hooks must never break the host session: warn and use defaults.
		fmt.Fprintln(os.Stderr, "config warning: "+err.Error())
		cfg = &config.Config{}
	}
	wfs := loadActiveWorkflows()
	cwd, _ := os.Getwd()
	d := evalPrewrite(filePath, cwd, cfg, wfs)
	if d.Allow {
		return nil
	}
	if d.NeedInit {
		telemetry.RecordBlock(cfg, "", "", string(d.FileType), filePath, "need-init")
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), "", "—", filePath))
		fmt.Fprintln(os.Stderr, ui.StyleMuted.Render("Run: centinela start <feature>"))
		exitPrewrite(2)
	}
	telemetry.RecordBlock(cfg, d.Feature, d.Step, string(d.FileType), filePath, "out-of-step")
	fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), d.Step, d.Feature, filePath))
	exitPrewrite(2)
	return nil
}
