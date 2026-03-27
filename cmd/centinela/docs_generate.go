package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/docgen"
	"github.com/samuelnp/centinela/internal/ui"
)

var docsOut string
var docsTitle string

var docsGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate HTML documentation from Centinela artifacts",
	RunE:  runDocsGenerate,
}

func init() {
	docsGenerateCmd.Flags().StringVar(&docsOut, "out", "docs/project-docs/index.html", "output html file path")
	docsGenerateCmd.Flags().StringVar(&docsTitle, "title", "Centinela Project Documentation", "report title")
	docsCmd.AddCommand(docsGenerateCmd)
}

func runDocsGenerate(_ *cobra.Command, _ []string) error {
	if err := docgen.Generate(docsOut, docsTitle); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Documentation generated: " + docsOut))
	return nil
}
