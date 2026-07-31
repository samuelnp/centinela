package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var routeSetReason string

var routeSetCmd = &cobra.Command{
	Use:   "set <feature> <role> <tier>",
	Short: "Record the model tier a role runs at for this feature",
	Long: "Records role→tier into the feature's workflow state after the refusal\n" +
		"matrix passes: floors are enforced (the error names the floor), a tier\n" +
		"below the static default demands --reason, and a downgrade is refused\n" +
		"once the role's step is underway. Upgrades are allowed anytime.",
	Args:          cobra.ExactArgs(3),
	RunE:          runRouteSet,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	routeSetCmd.Flags().StringVar(&routeSetReason, "reason", "", "why this tier (required below the static default)")
	routeCmd.AddCommand(routeSetCmd)
}

func runRouteSet(_ *cobra.Command, args []string) error {
	feature := args[0]
	cfg, err := loadRoutingConfig()
	if err != nil {
		return err
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}
	role, tier, err := orchestration.ParseRouteTarget(wf.CurrentStep == "done", args[1], args[2])
	if err != nil {
		return err
	}
	req := buildRouteRequest(wf, cfg, role, tier, routeSetReason)
	if err := orchestration.ValidateRoute(req); err != nil {
		return err
	}
	wf.SetModelRoute(string(role), workflow.ModelRoute{
		Tier:      string(tier),
		Reason:    routeSetReason,
		DecidedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err := workflow.Save(wf); err != nil {
		return fmt.Errorf("cannot save workflow: %w", err)
	}
	telemetry.RecordRouteDecision(cfg, feature, string(role), string(tier),
		string(req.CurrentTier), routeSetReason, resolveEmitModel(wf, cfg))
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("route set: %s → %s (was %s)", role, tier, req.CurrentTier)))
	return nil
}
