package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/gitdiff"
	"github.com/samuelnp/centinela/internal/githooks"
)

var precommitInstallCmd = &cobra.Command{
	Use:           "install",
	Short:         "Install the centinela pre-commit git hook (marker-delimited, no clobber)",
	RunE:          runPrecommitInstall,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var precommitUninstallCmd = &cobra.Command{
	Use:           "uninstall",
	Short:         "Remove the centinela block from the pre-commit git hook",
	RunE:          runPrecommitUninstall,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	precommitCmd.AddCommand(precommitInstallCmd)
	precommitCmd.AddCommand(precommitUninstallCmd)
}

func runPrecommitInstall(cmd *cobra.Command, _ []string) error {
	dir := hooksDir()
	changed, err := githooks.Install(dir)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "pre-commit")
	if changed {
		fmt.Fprintf(cmd.OutOrStdout(), "Installed centinela pre-commit hook at %s\n", path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "centinela pre-commit hook already installed at %s\n", path)
	}
	return nil
}

func runPrecommitUninstall(cmd *cobra.Command, _ []string) error {
	dir := hooksDir()
	changed, err := githooks.Uninstall(dir)
	if err != nil {
		return err
	}
	if changed {
		fmt.Fprintln(cmd.OutOrStdout(), "Removed centinela pre-commit hook block")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "No centinela pre-commit hook block to remove")
	}
	return nil
}

// hooksDir resolves the repo's hooks directory via git, falling back to
// .git/hooks when git is unavailable.
func hooksDir() string {
	out, err := gitdiff.Default.Run("git", "rev-parse", "--git-path", "hooks")
	if err != nil {
		return filepath.Join(".git", "hooks")
	}
	if dir := strings.TrimSpace(out); dir != "" {
		return dir
	}
	return filepath.Join(".git", "hooks")
}
