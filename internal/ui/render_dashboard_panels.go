package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/teamdashboard"
)

// dashSummaryLabel is the one-line header summary: active feature count and
// roadmap presence.
func dashSummaryLabel(d teamdashboard.Dashboard) string {
	road := "no roadmap"
	if d.Roadmap.Present {
		road = fmt.Sprintf("%d/%d roadmap done", d.Roadmap.Done, d.Roadmap.Total)
	}
	return fmt.Sprintf("%d active · %s", len(d.Features), road)
}

// featuresPanel renders Panel 1. Empty Features prints an honest empty state.
func featuresPanel(rows []teamdashboard.FeatureRow) string {
	lines := []string{StyleBold.Render("In-flight features")}
	if len(rows) == 0 {
		lines = append(lines, StyleMuted.Render("  no active features — run `centinela start <feature>`"))
		return strings.Join(lines, "\n")
	}
	for _, r := range rows {
		lines = append(lines, featureRow(r))
	}
	return strings.Join(lines, "\n")
}

func featureRow(r teamdashboard.FeatureRow) string {
	return fmt.Sprintf("  %s  %s  %s  %s",
		StyleBold.Render(r.Feature),
		fmt.Sprintf("%s %d/%d", r.Step, r.StepIndex, r.StepTotal),
		StyleMuted.Render(fmt.Sprintf("%dd · %s · %s · %s",
			r.AgeDays, orDefault(r.Profile, "default"),
			orDefault(r.Archetype, "canonical"), orDefault(r.Worktree, "—"))),
		StyleMuted.Render("owner "+r.Owner))
}

// roadmapPanel renders Panel 2. A non-present roadmap prints an empty state;
// an empty-but-present roadmap renders "0/0 done".
func roadmapPanel(rb teamdashboard.RoadmapBurndown) string {
	lines := []string{StyleBold.Render("Roadmap burn-down")}
	if !rb.Present {
		lines = append(lines, StyleMuted.Render("  no roadmap — run `centinela roadmap …`"))
		return strings.Join(lines, "\n")
	}
	for _, p := range rb.Phases {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			p.Name, StyleMuted.Render(fmt.Sprintf("%d/%d", p.Done, p.Total))))
	}
	lines = append(lines, StyleBold.Render(fmt.Sprintf("  %d/%d done", rb.Done, rb.Total)))
	return strings.Join(lines, "\n")
}

// gatesPanel renders Panel 3. No gate failures prints an honest empty state.
func gatesPanel(gates []teamdashboard.GateHealth) string {
	lines := []string{StyleBold.Render("Gate health")}
	if len(gates) == 0 {
		lines = append(lines, StyleMuted.Render("  no gate failures recorded"))
		return strings.Join(lines, "\n")
	}
	for _, g := range gates {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			StyleMuted.Render(fmt.Sprintf("%3d", g.Fails)), g.Gate))
	}
	return strings.Join(lines, "\n")
}

// orDefault returns def when s is empty.
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
