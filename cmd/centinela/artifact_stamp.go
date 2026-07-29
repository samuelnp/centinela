package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/treestate"
	"github.com/samuelnp/centinela/internal/workflow"
)

// artifactStampCmd is the verifier's LAST action: it binds the report it just
// wrote to the tree state it verified. Domain logic lives in internal/workflow
// per G7 — this file ONLY wires the Cobra surface.
var artifactStampCmd = &cobra.Command{
	Use:   "stamp <feature>",
	Short: "Record the verified revision and tree digest in the gatekeeper report",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactStamp,
}

func init() {
	artifactCmd.AddCommand(artifactStampCmd)
}

func runArtifactStamp(_ *cobra.Command, args []string) error {
	feature := args[0]
	if err := requireKnownFeature(feature); err != nil {
		return err
	}
	snapshot, err := workflow.StampVerification(feature, ".", treestate.NewExecRunner())
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "stamped %s\n  revision:   %s\n  treeDigest: %s\n",
		workflow.GatekeeperReportPath(feature), snapshot.Revision, snapshot.Digest)
	return nil
}
