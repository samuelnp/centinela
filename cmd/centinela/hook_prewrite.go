package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/hookpolicy"
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

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	wfs := loadActiveWorkflows()
	cwd, _ := os.Getwd()
	d := hookpolicy.EvaluatePrewrite(filePath, cwd, cfg, wfs)
	if d.Allow {
		return nil
	}
	if d.NeedInit {
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), "", "—", filePath))
		fmt.Fprintln(os.Stderr, ui.StyleMuted.Render("Run: centinela start <feature>"))
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), d.Step, d.Feature, filePath))
	os.Exit(2)
	return nil
}
