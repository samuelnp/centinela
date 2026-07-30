package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var routeShowCmd = &cobra.Command{
	Use:           "show <feature>",
	Short:         "Show the effective role→tier table for a feature",
	Long:          "Renders every role this workflow's steps schedule with its effective\ntier, whether it came from a route or from static config, its floor,\nthe recorded reason, and when it was decided.",
	Args:          cobra.ExactArgs(1),
	RunE:          runRouteShow,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	routeCmd.AddCommand(routeShowCmd)
}

func runRouteShow(_ *cobra.Command, args []string) error {
	feature := args[0]
	cfg, err := loadRoutingConfig()
	if err != nil {
		return err
	}
	wf, err := workflow.Load(feature)
	if err != nil {
		return err
	}
	models, _ := orchestrationRouting(cfg)
	floors := config.OrchestrationFloors(cfg)
	rows := routeRows(wf, scheduledRoles(wf), models, floors)
	hint := orchestration.RoutingDirective(feature,
		workflow.RequiredEvidenceRoles(feature, wf.CurrentStep),
		workflow.RouteTiers(wf), floors, models)
	fmt.Println(ui.RenderRouteTable(feature, rows, hint))
	return nil
}

// scheduledRoles collects the roles this workflow's own steps require, in step
// order, deduped — the same contract-aware set `route set` accepts.
func scheduledRoles(wf *workflow.Workflow) []orchestration.Role {
	var roles []orchestration.Role
	seen := map[orchestration.Role]bool{}
	for _, step := range wf.OrderedSteps() {
		for _, role := range workflow.RequiredEvidenceRoles(wf.Feature, step) {
			if !seen[role] {
				seen[role] = true
				roles = append(roles, role)
			}
		}
	}
	return roles
}

// routeRows projects each scheduled role onto a presentation row. Pure shuffling:
// the tier, floor, and route facts are all resolved by the domain (G7).
func routeRows(wf *workflow.Workflow, roles []orchestration.Role, models orchestration.RoleModels, floors map[string]string) []ui.RouteRow {
	rows := make([]ui.RouteRow, 0, len(roles))
	for _, role := range roles {
		row := ui.RouteRow{Role: string(role), Source: "static", Tier: string(orchestration.RoleTier(role, models))}
		if route, ok := wf.ModelRouteFor(string(role)); ok {
			row.Reason, row.DecidedAt = route.Reason, route.DecidedAt
			// The table must report what the hook EMITS, never the raw stored
			// string: a corrupt or below-floor route is ignored by the overlay,
			// so it renders as the static tier flagged "ignored" rather than
			// masquerading as the effective one.
			if tier, honored := orchestration.HonoredRouteTier(role, route.Tier, floors); honored {
				row.Source, row.Tier = "routed", string(tier)
			} else {
				row.Source = "ignored"
			}
		}
		if floor, ok := orchestration.EffectiveFloor(role, floors); ok {
			row.Floor = string(floor)
		}
		rows = append(rows, row)
	}
	return rows
}
