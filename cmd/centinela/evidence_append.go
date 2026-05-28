package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

var evidenceAppendCmd = &cobra.Command{
	Use:   "append <feature> <role> <field> <value>",
	Short: "Append to a list field (inputs|outputs|edgeCases); dedup on exact match",
	Args:  cobra.ExactArgs(4),
	RunE:  runEvidenceAppend,
}

func init() {
	evidenceCmd.AddCommand(evidenceAppendCmd)
}

func runEvidenceAppend(_ *cobra.Command, args []string) error {
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
	if err := evidence.AppendField(doc, field, value); err != nil {
		return err
	}
	return evidence.WriteAtomic(feature, role, doc)
}
