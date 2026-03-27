package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/ui"
)

var hookStatuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Hook: render compact Centinela status line",
	RunE:  runHookStatusline,
}

func init() {
	hookCmd.AddCommand(hookStatuslineCmd)
}

func runHookStatusline(_ *cobra.Command, _ []string) error {
	io.ReadAll(os.Stdin) //nolint:errcheck // stdin is optional for this hook
	view := buildStatusLineView(loadActiveWorkflows())
	out := ui.RenderStatusLine(view)
	if out != "" {
		fmt.Println(out)
	}
	return nil
}
