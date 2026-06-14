package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// roadmapMarkdownFile is the human-readable roadmap generated from roadmap.json.
const roadmapMarkdownFile = "ROADMAP.md"

var roadmapGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate ROADMAP.md deterministically from .workflow/roadmap.json",
	RunE:  runRoadmapGenerate,
}

func init() {
	roadmapCmd.AddCommand(roadmapGenerateCmd)
}

// runRoadmapGenerate is a thin orchestrator: load the roadmap, render markdown
// in the roadmap domain, and write the file. No formatting logic lives here.
func runRoadmapGenerate(_ *cobra.Command, _ []string) error {
	r, err := roadmap.Load()
	if err != nil {
		return roadmapCommandError(err)
	}
	data := roadmap.RenderMarkdown(r)
	if err := os.WriteFile(roadmapMarkdownFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s\n", roadmapMarkdownFile)
	return nil
}
