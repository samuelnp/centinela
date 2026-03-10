package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run built-in gate checks and all validate commands from centinela.toml",
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, _ []string) error {
	return executeValidation()
}

// executeValidation is shared by runValidate and runComplete (validate step).
func executeValidation() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	allPassed := true

	// --- Built-in gates ---
	fmt.Println(ui.StyleBold.Render("Built-in Gates"))
	results := gates.RunAll(cfg)
	for _, r := range results {
		fmt.Println(ui.RenderGateResult(r))
	}
	if !gates.AllPassed(results) {
		allPassed = false
	}

	// --- User commands ---
	if len(cfg.Validate.Commands) > 0 {
		fmt.Println()
		fmt.Println(ui.StyleBold.Render("Validate Commands"))
		for _, cmd := range cfg.Validate.Commands {
			passed, out := runCommand(cmd)
			fmt.Println(ui.RenderCmdResult(cmd, passed, out))
			if !passed {
				allPassed = false
			}
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println(ui.RenderSuccess("All gates passed."))
		return nil
	}
	return fmt.Errorf("validation failed — fix the issues above before completing the validate step")
}

// runCommand executes a shell command and returns (passed, combined output).
func runCommand(cmdStr string) (bool, string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return err == nil, strings.TrimSpace(buf.String())
}
