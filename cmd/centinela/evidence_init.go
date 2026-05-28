package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/workflow"
)

var evidenceInitCmd = &cobra.Command{
	Use:   "init <feature> <role>",
	Short: "Drop a schema-valid JSON + companion markdown skeleton for a role",
	Args:  cobra.ExactArgs(2),
	RunE:  runEvidenceInit,
}

func init() {
	evidenceCmd.AddCommand(evidenceInitCmd)
}

func runEvidenceInit(_ *cobra.Command, args []string) error {
	feature, roleArg := args[0], args[1]
	if err := requireKnownFeature(feature); err != nil {
		return err
	}
	role, err := evidence.ParseRole(roleArg)
	if err != nil {
		return err
	}
	release, err := evidence.Lock(feature, role)
	if err != nil {
		return err
	}
	defer release()
	skel := evidence.Skeleton(feature, role, Version)
	if err := evidence.WriteAtomic(feature, role, skel); err != nil {
		return err
	}
	if err := evidence.WriteCompanion(feature, role, evidence.DefaultCompanionTemplate(feature, role)); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "wrote %s and companion .md\n", featureRolePath(feature, role))
	return nil
}

// requireKnownFeature returns an error if no .workflow/<feature>.json exists.
// Lists active features in the error so the agent can spot a typo fast.
func requireKnownFeature(feature string) error {
	if _, err := os.Stat(workflow.FilePath(feature)); err == nil {
		return nil
	}
	active := workflow.ActiveWorkflows(workflow.WorkflowDir)
	names := make([]string, 0, len(active))
	for _, wf := range active {
		names = append(names, wf.Feature)
	}
	return fmt.Errorf("unknown feature %q (active: %v) — run `centinela start %s` first", feature, names, feature)
}

func featureRolePath(feature string, role evidence.Role) string {
	return fmt.Sprintf(".workflow/%s-%s.json", feature, role)
}
