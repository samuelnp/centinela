package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
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

	cwd, err := os.Getwd()
	if err == nil {
		rel, relErr := filepath.Rel(cwd, filePath)
		if relErr != nil || strings.HasPrefix(rel, "..") {
			return nil
		}
	}

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	fileType := workflow.ClassifyFile(filePath, cfg)
	if fileType == workflow.TypeOther || fileType == workflow.TypeRoadmap {
		return nil
	}

	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(fileType), "", "—", filePath))
		fmt.Fprintln(os.Stderr, ui.StyleMuted.Render("Run: centinela start <feature>"))
		os.Exit(2)
	}

	for _, path := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(path), ".json"))
		if err != nil {
			continue
		}
		if wf.CurrentStep == "done" || workflow.IsAllowedInStep(fileType, wf.CurrentStep) {
			return nil
		}
	}

	wf, _ := workflow.Load(strings.TrimSuffix(filepath.Base(entries[0]), ".json"))
	if wf != nil {
		fmt.Fprintln(os.Stderr, ui.RenderBlocked(string(fileType), wf.CurrentStep, wf.Feature, filePath))
	}
	os.Exit(2)
	return nil
}
