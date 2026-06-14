package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/doctor"
	"github.com/samuelnp/centinela/internal/ui"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose (and, with --fix, safely repair) Centinela project health",
	Long: "Runs read-only health checks (hooks, roadmap, worktrees, workflow " +
		"state, evidence, config, version). With --fix, applies only safe, " +
		"idempotent repairs; destructive actions are reported, never applied.",
	RunE:          runDoctor,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Apply safe, idempotent repairs (never destructive)")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(_ *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	ctx, err := doctor.NewContext(cwd)
	if err != nil {
		return err
	}
	var diags []doctor.Diagnosis
	if doctorFix {
		diags = doctor.Fix(ctx)
	} else {
		diags = doctor.Run(ctx)
	}
	for _, d := range diags {
		fmt.Fprintln(os.Stdout, ui.RenderDiagnosis(d))
	}
	fmt.Fprintln(os.Stdout, ui.RenderDoctorSummary(diags))
	if doctor.ExitError(diags) {
		return errDoctorFailed
	}
	return nil
}

// errDoctorFailed drives exit code 1 when any check is ERROR. The diagnosis
// report on stdout is the contract; SilenceErrors keeps cobra from echoing
// this sentinel, and main prints it to stderr (separate from the report).
var errDoctorFailed = errors.New("centinela doctor: one or more checks reported errors")
