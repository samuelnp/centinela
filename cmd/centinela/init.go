package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/scaffold"
	"github.com/samuelnp/centinela/internal/setup"
	"github.com/samuelnp/centinela/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold docs and wire centinela hooks into .claude/settings.json",
	Long: "Creates CLAUDE.md, PROJECT.md.template, and docs/architecture/ from\n" +
		"embedded templates, then merges centinela hooks into the Claude settings\n" +
		"file. Safe to run multiple times — existing files are never overwritten.",
	RunE: runInit,
}

var localFlag bool

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&localFlag, "local", false,
		"Write hooks to .claude/settings.local.json instead of settings.json")
}

func runInit(_ *cobra.Command, _ []string) error {
	result, err := scaffold.Extract(".")
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}
	for _, f := range result.Created {
		fmt.Println(ui.RenderSuccess(f))
	}
	for _, f := range result.Skipped {
		fmt.Println(ui.StyleMuted.Render("  skipped  " + f))
	}
	if len(result.Created) > 0 {
		fmt.Println()
	}

	settingsPath := ".claude/settings.json"
	if localFlag {
		settingsPath = ".claude/settings.local.json"
	}
	changed, err := setup.InjectHooks(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", settingsPath, err)
	}

	if !changed {
		fmt.Println(ui.StyleMuted.Render("hooks already present in " + settingsPath))
		return nil
	}
	fmt.Println(ui.RenderSuccess("hooks wired in " + settingsPath))
	fmt.Println(ui.StyleMuted.Render("  PreToolUse  (Write, Edit)  →  centinela hook prewrite"))
	fmt.Println(ui.StyleMuted.Render("  PostToolUse (Write, Edit)  →  centinela hook postwrite"))
	fmt.Println(ui.StyleMuted.Render("  UserPromptSubmit            →  centinela hook context"))
	return nil
}
