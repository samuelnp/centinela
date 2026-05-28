package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceReadField string

var evidenceReadCmd = &cobra.Command{
	Use:   "read <feature> <role>",
	Short: "Print the evidence doc (or one field) as JSON to stdout",
	Args:  cobra.ExactArgs(2),
	RunE:  runEvidenceRead,
}

func init() {
	evidenceReadCmd.Flags().StringVar(&evidenceReadField, "field", "", "single field to read (e.g. outputs, extra.note)")
	evidenceCmd.AddCommand(evidenceReadCmd)
}

func runEvidenceRead(_ *cobra.Command, args []string) error {
	feature, roleArg := args[0], args[1]
	role, err := evidence.ParseRole(roleArg)
	if err != nil {
		return err
	}
	doc, err := evidence.Read(feature, role)
	if err != nil {
		if evidence.IsNotFound(err) {
			return fmt.Errorf("%w — run `centinela evidence init %s %s` first", err, feature, role)
		}
		return err
	}
	if evidenceReadField == "" {
		out, err := doc.MarshalJSON()
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(out)
		return err
	}
	v, err := evidence.ReadField(doc, evidenceReadField)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
