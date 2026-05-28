package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceSetCmd = &cobra.Command{
	Use:   "set <feature> <role> <field> <value>",
	Short: "Set a scalar field atomically (supports extra.<key> for free-form)",
	Args:  cobra.ExactArgs(4),
	RunE:  runEvidenceSet,
}

func init() {
	evidenceCmd.AddCommand(evidenceSetCmd)
}

func runEvidenceSet(_ *cobra.Command, args []string) error {
	feature, roleArg, field, value := args[0], args[1], args[2], args[3]
	role, err := evidence.ParseRole(roleArg)
	if err != nil {
		return err
	}
	release, err := evidence.Lock(feature, role)
	if err != nil {
		return err
	}
	defer release()
	doc, err := evidence.Read(feature, role)
	if err != nil {
		if evidence.IsNotFound(err) {
			return fmt.Errorf("%w — run `centinela evidence init %s %s` first", err, feature, role)
		}
		return err
	}
	if err := evidence.SetField(doc, field, value); err != nil {
		return err
	}
	return evidence.WriteAtomic(feature, role, doc)
}
