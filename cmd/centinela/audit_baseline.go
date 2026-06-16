package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
)

// auditBaselineCmd records (or replaces) the committed ratchet snapshot. It
// full-scans the participating gates, fingerprints the current Fail violations,
// and writes the deterministic baseline file. It never re-adds resolved
// violations (the file is fully replaced), so the ratchet only tightens.
var auditBaselineCmd = &cobra.Command{
	Use:           "baseline",
	Short:         "Record the current gate violations as the tolerated baseline",
	RunE:          runAuditBaseline,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	auditCmd.AddCommand(auditBaselineCmd)
}

func runAuditBaseline(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	path := cfg.Gates.AuditBaseline.BaselinePath
	b := audit.Record(cfg)
	if err := audit.Save(path, b); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s — %d gate(s), %d violation(s) baselined\n",
		path, len(b.Gates), countFingerprints(b))
	return nil
}

func countFingerprints(b audit.Baseline) int {
	n := 0
	for _, e := range b.Gates {
		n += len(e.Fingerprints)
	}
	return n
}
