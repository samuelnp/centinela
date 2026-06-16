package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/audit"
)

// auditVerdict is the machine-readable shape emitted by `audit --json`, so a
// caller (e.g. precommit-and-pr-gate) reads counts + the new fingerprints
// without scraping text. Exit code still encodes the pass/fail verdict.
type auditVerdict struct {
	New       int                 `json:"new"`
	Baselined int                 `json:"baselined"`
	Resolved  int                 `json:"resolved"`
	NewItems  []audit.Fingerprint `json:"new_items"`
}

// printAuditJSON marshals the diff verdict and returns a non-nil error iff there
// are new violations, so the JSON path keeps the same exit semantics as text.
func printAuditJSON(cmd *cobra.Command, d audit.Diff) error {
	v := auditVerdict{
		New:       len(d.New),
		Baselined: len(d.Baselined),
		Resolved:  len(d.Resolved),
		NewItems:  d.New,
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	if d.HasNew() {
		return fmt.Errorf("%d new violation(s) since baseline", len(d.New))
	}
	return nil
}

// renderNoBaseline reports the safe-adoption default: no baseline recorded yet
// is non-blocking (exit 0), in both text and JSON form.
func renderNoBaseline(cmd *cobra.Command) error {
	if auditJSON {
		data, _ := json.MarshalIndent(auditVerdict{NewItems: []audit.Fingerprint{}}, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}
	fmt.Fprintln(cmd.OutOrStdout(), "no baseline — run `centinela audit baseline`")
	return nil
}
