package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
)

var readyJSON bool

var roadmapReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List features that are ready to start right now (all dependencies done)",
	RunE:  runRoadmapReady,
}

func init() {
	roadmapReadyCmd.Flags().BoolVar(&readyJSON, "json", false, "Emit the ready feature names as a JSON array")
	roadmapCmd.AddCommand(roadmapReadyCmd)
}

func runRoadmapReady(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	ready := roadmap.ReadySet(r)
	if readyJSON {
		if ready == nil {
			ready = []string{}
		}
		data, err := json.MarshalIndent(ready, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	fmt.Println(ui.RenderReadyList(ready))
	return nil
}
