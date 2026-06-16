package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/ui"
)

var auditJSON bool

// auditCmd is the parent for the baseline + ratchet command group. Its bare
// RunE is the ratchet check: full-scan the participating gates, diff against the
// committed baseline, render, and exit non-zero iff any new violation appears.
// A missing baseline is non-blocking (safe adoption default).
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Ratchet current gate violations against the committed baseline",
	Long: "Compares current full-scan gate violations against\n" +
		".workflow/audit-baseline.json. New violations block (exit 1);\n" +
		"baselined ones are tolerated; resolved ones are reported as prunable.\n" +
		"Record a baseline with `centinela audit baseline`.",
	RunE:          runAudit,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "Emit the ratchet verdict as JSON")
	rootCmd.AddCommand(auditCmd)
}

func runAudit(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	b, exists, err := audit.Load(cfg.Gates.AuditBaseline.BaselinePath)
	if err != nil {
		return err
	}
	if !exists {
		return renderNoBaseline(cmd)
	}
	d := audit.Ratchet(cfg, b)
	if auditJSON {
		return printAuditJSON(cmd, d)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.RenderAuditDiff(d))
	if d.HasNew() {
		return fmt.Errorf("%d new violation(s) since baseline", len(d.New))
	}
	return nil
}
