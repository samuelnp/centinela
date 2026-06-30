package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
	"github.com/samuelnp/centinela/internal/worktree"
)

var hookPostwriteCmd = &cobra.Command{
	Use:   "postwrite",
	Short: "Hook: inject workflow tag after every file write",
	RunE:  runHookPostwrite,
}

func init() {
	hookCmd.AddCommand(hookPostwriteCmd)
}

type postwriteInput struct {
	ToolInput struct {
		FilePath  string `json:"file_path"`
		FilePath2 string `json:"filePath"`
		Command   string `json:"command"`
	} `json:"tool_input"`
}

func runHookPostwrite(_ *cobra.Command, _ []string) error {
	raw, _ := io.ReadAll(os.Stdin)
	reformatPostwrite(raw)
	entries, _ := filepath.Glob(filepath.Join(workflow.WorkflowDir, "*.json"))
	for _, path := range entries {
		wf, err := workflow.Load(strings.TrimSuffix(filepath.Base(path), ".json"))
		if err != nil {
			continue
		}
		fmt.Println(ui.RenderTag(wf))
	}
	return nil
}

// reformatPostwrite parses the hook payload, identifies the written file,
// and rewrites it via hookpolicy.FormatEvidence when it is the active
// feature's evidence JSON. Errors are swallowed — the postwrite hook is
// best-effort and must never block the user.
func reformatPostwrite(raw []byte) {
	path := extractPostwritePath(raw)
	if path == "" {
		return
	}
	feature, _ := worktree.DetectFeatureFromCwd(mustGetwd())
	body, err := os.ReadFile(path)
	if err != nil {
		return
	}
	out, changed, err := hookpolicy.FormatEvidence(path, body, feature)
	if err != nil || !changed {
		return
	}
	_ = evidence.WriteBytesAtomic(path, out)
}

func extractPostwritePath(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var in postwriteInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return ""
	}
	if in.ToolInput.FilePath != "" {
		return in.ToolInput.FilePath
	}
	if in.ToolInput.FilePath2 != "" {
		return in.ToolInput.FilePath2
	}
	if paths := hookpolicy.ExtractApplyPatchPaths(in.ToolInput.Command); len(paths) > 0 {
		return paths[0]
	}
	return ""
}

func mustGetwd() string {
	d, _ := os.Getwd()
	return d
}
