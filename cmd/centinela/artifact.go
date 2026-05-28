package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/evidence"
)

// artifactCmd is the top-level orchestrator for `centinela artifact`. Domain
// logic — template rendering, path scoping, atomic write — lives in
// internal/evidence per G7. This file ONLY wires the Cobra surface.
var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Drop pre-filled .workflow/<feature>-<kind> stubs (edge-cases, gatekeeper, etc.)",
}

var artifactForce bool

var artifactNewCmd = &cobra.Command{
	Use:   "new <feature> <kind>",
	Short: "Create the templated artifact for kind under .workflow/",
	Args:  cobra.ExactArgs(2),
	RunE:  runArtifactNew,
}

func init() {
	artifactNewCmd.Flags().BoolVar(&artifactForce, "force", false, "overwrite if the artifact already exists")
	artifactCmd.AddCommand(artifactNewCmd)
	rootCmd.AddCommand(artifactCmd)
}

func runArtifactNew(_ *cobra.Command, args []string) error {
	feature, kindArg := args[0], args[1]
	if err := requireKnownFeature(feature); err != nil {
		return err
	}
	kind, err := evidence.ParseKind(kindArg)
	if err != nil {
		return err
	}
	paths, err := evidence.WriteArtifact(feature, kind, artifactForce)
	if err != nil {
		return err
	}
	for _, p := range paths {
		fmt.Fprintf(os.Stdout, "wrote %s\n", p)
	}
	return nil
}
