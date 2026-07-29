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
	// Markers live in the primary tree, so the hook must look there whatever
	// the session CWD is. Outside a repo, fall back to the CWD.
	repo, err := worktree.PrimaryTree(".")
	if err != nil {
		repo = "."
	}
	validate := stewardEvidenceValidatorFor(repo)
	markers, _ := filepath.Glob(filepath.Join(repo, ".workflow", "*-merge-pending.json"))
	for _, p := range markers {
		feature := strings.TrimSuffix(filepath.Base(p), "-merge-pending.json")
		m, err := worktree.LoadPending(repo, feature)
		if err != nil || m == nil {
			continue
		}
		if _, err := validate(feature); err == nil {
			continue // valid steward evidence — stop re-emitting
		}
		fmt.Println(ui.RenderMergeStewardNeeded(m.Feature, m.Reason))
		fmt.Println(m.Directive())
	}
	return nil
}
