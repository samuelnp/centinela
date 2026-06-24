package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/audit"
)

// adoptVerdict is the machine-readable shape emitted by `adopt --json`, so the
// onboarding agent reads the verdict + counts without scraping text. Exit code
// still encodes adopted (0) vs skipped (non-zero).
type adoptVerdict struct {
	Adopted bool           `json:"adopted"`
	Skipped bool           `json:"skipped"`
	Path    string         `json:"path"`
	Total   int            `json:"total"`
	PerGate map[string]int `json:"per_gate"`
}

// printAdoptJSON marshals the adoption verdict and returns a non-nil error iff the
// adoption was skipped, so the JSON path keeps the same exit semantics as text.
func printAdoptJSON(cmd *cobra.Command, o audit.Outcome) error {
	v := adoptVerdict{
		Adopted: !o.Skipped,
		Skipped: o.Skipped,
		Path:    o.Path,
		Total:   o.Baseline.Total(),
		PerGate: perGateCounts(o.Baseline),
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	if o.Skipped {
		return fmt.Errorf("baseline already exists at %s — use --force to overwrite", o.Path)
	}
	return nil
}

// perGateCounts maps each gate name to its accepted-finding count. It is an empty
// (non-nil) map when the baseline is zero value, so the JSON renders {} not null.
func perGateCounts(b audit.Baseline) map[string]int {
	out := make(map[string]int, len(b.Gates))
	for _, e := range b.Gates {
		out[e.Gate] = len(e.Fingerprints)
	}
	return out
}
