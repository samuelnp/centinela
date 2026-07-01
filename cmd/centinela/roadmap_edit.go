package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	editName        string
	editDescription string
	editArchetype   string
	editDependsOn   []string
)

var roadmapEditCmd = &cobra.Command{
	Use:     "edit <slug>",
	Aliases: []string{"update"},
	Short:   "Edit a feature in place (rename rewrites every dependent)",
	Args:    cobra.ExactArgs(1),
	RunE:    runRoadmapEdit,
}

func init() {
	roadmapEditCmd.Flags().StringVar(&editName, "name", "", "new slug (rename; rewrites dependents)")
	roadmapEditCmd.Flags().StringVar(&editDescription, "description", "", "replace the human-facing bullet prose")
	roadmapEditCmd.Flags().StringVar(&editArchetype, "archetype", "", "replace the workflow archetype")
	roadmapEditCmd.Flags().StringSliceVar(&editDependsOn, "depends-on", nil, "replace the dependsOn list")
	roadmapCmd.AddCommand(roadmapEditCmd)
}

func runRoadmapEdit(cmd *cobra.Command, args []string) error {
	if err := roadmap.Edit(roadmap.RoadmapFile, roadmap.EditRequest{
		Slug:        args[0],
		NewName:     editName,
		Description: editDescription,
		Archetype:   editArchetype,
		DependsOn:   editDependsOn,
		SetDeps:     cmd.Flags().Changed("depends-on"),
	}); err != nil {
		return err
	}
	fmt.Println(ui.RenderSuccess(fmt.Sprintf(
		"Edited %q. Remember to sync ROADMAP.md (roadmap-doc-sync).", args[0])))
	return nil
}
