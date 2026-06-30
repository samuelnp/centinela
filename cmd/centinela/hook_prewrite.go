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
	"github.com/samuelnp/centinela/internal/workflow"
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
		Command   string `json:"command"`
	} `json:"tool_input"`
}

var exitPrewrite = os.Exit
var evalPrewriteMulti = hookpolicy.EvaluatePrewriteMulti

// prewriteTargets resolves the paths a write touches. Claude/OpenCode send
// file_path/filePath; Codex's apply_patch sends a patch envelope in command,
// which may touch several files.
func prewriteTargets(in prewriteInput) []string {
	if p := in.ToolInput.FilePath; p != "" {
		return []string{p}
	}
	if p := in.ToolInput.FilePath2; p != "" {
		return []string{p}
	}
	return hookpolicy.ExtractApplyPatchPaths(in.ToolInput.Command)
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
	paths := prewriteTargets(input)
	if len(paths) == 0 {
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
	d := evalPrewriteMulti(paths, cwd, cfg, wfs)
	if d.Allow {
		return nil
	}
	blockPrewrite(d, cfg, wfs)
	return nil
}

func blockPrewrite(d hookpolicy.PrewriteDecision, cfg *config.Config, wfs []*workflow.Workflow) {
	model := resolveEmitModelFrom(wfs, cfg)
	if d.NeedInit {
		telemetry.RecordBlock(cfg, "", "", string(d.FileType), d.Path, "need-init", model)
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), "", "—", d.Path))
		fmt.Fprintln(os.Stderr, ui.StyleMuted.Render("Run: centinela start <feature>"))
		exitPrewrite(2)
		return
	}
	telemetry.RecordBlock(cfg, d.Feature, d.Step, string(d.FileType), d.Path, "out-of-step", model)
	fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(d.FileType), d.Step, d.Feature, d.Path))
	exitPrewrite(2)
}
