package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/docgen"
	"github.com/samuelnp/centinela/internal/ui"
)

var docsValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate inputs required for docs generation",
	RunE:  runDocsValidate,
}

func init() {
	docsCmd.AddCommand(docsValidateCmd)
}

func runDocsValidate(_ *cobra.Command, _ []string) error {
	if err := docgen.ValidateInputs(); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess("Documentation inputs are valid."))
	return nil
}
