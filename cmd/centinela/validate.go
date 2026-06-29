package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

var (
	validateChanged bool
	validateFull    bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run built-in gate checks and all validate commands from centinela.toml",
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().BoolVar(&validateChanged, "changed", false, "Run built-in gates only over files changed since the diff base")
	validateCmd.Flags().BoolVar(&validateFull, "full", false, "Force a full-repo scan even when diff_mode would be diff-aware")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, _ []string) error {
	if validateChanged && validateFull {
		return fmt.Errorf("--changed and --full are mutually exclusive")
	}
	flag := config.FlagNone
	switch {
	case validateChanged:
		flag = config.FlagForceChanged
	case validateFull:
		flag = config.FlagForceFull
	}
	return executeValidationWithFlag(flag)
}

// executeValidation runs gates in the default mode (no CLI flag override).
// Kept as the package-level entry for complete.go and tests.
func executeValidation() error {
	return executeValidationWithFlag(config.FlagNone)
}

func executeValidationWithFlag(flag config.FlagOverride) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	mode := cfg.Validate.ResolveMode(currentEnv(), flag)
	filter, header := resolveDiffFilter(cfg, mode)

	allPassed := true
	fmt.Println(ui.StyleBold.Render("Built-in Gates " + header))
	results := appendAuditGate(cfg, gates.RunWithFilter(cfg, filter))
	emitGateFailures(cfg, results, resolveEmitModel(nil, cfg))
	for _, r := range results {
		fmt.Println(ui.RenderGateResult(r))
	}
	if !gates.AllPassed(results) {
		allPassed = false
	}

	if !runValidateCommands(cfg) {
		allPassed = false
	}

	emitCostWarning(cfg) // soft gate: surfaces an over-budget ⚠, never fails

	fmt.Println()
	if allPassed {
		fmt.Println(ui.RenderSuccess("All gates passed."))
		return nil
	}
	return fmt.Errorf("validation failed — fix the issues above before completing the validate step")
}

func runValidateCommands(cfg *config.Config) bool {
	if len(cfg.Validate.Commands) == 0 {
		return true
	}
	fmt.Println()
	fmt.Println(ui.StyleBold.Render("Validate Commands"))
	allPassed := true
	for _, cmd := range cfg.Validate.Commands {
		passed, out := runCommand(cmd)
		fmt.Println(ui.RenderCmdResult(cmd, passed, out))
		if !passed {
			allPassed = false
		}
	}
	return allPassed
}
