package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var roadmapShowJSON bool

var roadmapShowCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"list"},
	Short:   "Print the persisted roadmap as text, or --json for the verbatim Roadmap",
	RunE:    runRoadmapShow,
}

func init() {
	roadmapShowCmd.Flags().BoolVar(&roadmapShowJSON, "json", false, "Emit the persisted Roadmap verbatim as indented JSON")
	roadmapCmd.AddCommand(roadmapShowCmd)
}

func runRoadmapShow(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	if roadmapShowJSON {
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	fmt.Println(ui.RenderRoadmap(r))
	return nil
}
