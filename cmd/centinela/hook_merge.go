package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/worktree"
)

var hookMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Hook: re-emit the Merge Steward directive while a merge is pending",
	RunE:  runHookMerge,
}

func init() {
	hookCmd.AddCommand(hookMergeCmd)
}

func runHookMerge(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // drain stdin to avoid SIGPIPE
	markers, _ := filepath.Glob(filepath.Join(".workflow", "*-merge-pending.json"))
	for _, p := range markers {
		feature := strings.TrimSuffix(filepath.Base(p), "-merge-pending.json")
		m, err := worktree.LoadPending(".", feature)
		if err != nil || m == nil {
			continue
		}
		if _, err := stewardEvidenceValidator(feature); err == nil {
			continue // valid steward evidence — stop re-emitting
		}
		fmt.Println(ui.RenderMergeStewardNeeded(m.Feature, m.Reason))
		fmt.Println(m.Directive())
	}
	return nil
}
