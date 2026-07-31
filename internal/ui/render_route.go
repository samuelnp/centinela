package ui

import (
	"fmt"
	"strings"
)

// RouteRow is one already-resolved row of the effective routing table. Every
// value arrives as a rendered string: this package renders, it does not decide.
type RouteRow struct {
	Role      string
	Tier      string
	Source    string // "routed", "static", or "ignored" (recorded but not honored)
	Floor     string // "" when the role is floorless
	Reason    string
	DecidedAt string
}

// routeColumns are the table headers, in display order. The floor column is
// labelled "Route floor" on purpose: a floor bounds what a ROUTE may set, it
// does not raise a static tier the project configured — a bare "Floor" beside a
// lower static tier read as an enforcement claim the static path never makes.
var routeColumns = []string{"Role", "Tier", "Source", "Route floor", "Reason", "Decided"}

// RenderRouteTable renders the effective routing table for a feature. hint is
// the optional dynamic-routing directive appended below the table when roles are
// still un-routed; pass "" to omit it.
func RenderRouteTable(feature string, rows []RouteRow, hint string) string {
	if len(rows) == 0 {
		return StyleMuted.Render("no roles scheduled for " + feature)
	}
	cells := [][]string{routeColumns}
	for _, r := range rows {
		cells = append(cells, []string{r.Role, r.Tier, r.Source, dashIfEmpty(r.Floor),
			dashIfEmpty(r.Reason), dashIfEmpty(r.DecidedAt)})
	}
	widths := columnWidths(cells)
	lines := []string{StyleBold.Render("Model routing — " + feature), StyleBold.Render(routeLine(cells[0], widths))}
	for i, row := range cells[1:] {
		line := routeLine(row, widths)
		if rows[i].Source == "routed" {
			line = StyleGreen.Render(line)
		}
		lines = append(lines, line)
	}
	if hint != "" {
		lines = append(lines, "", StyleYellow.Render(hint))
	}
	return strings.Join(lines, "\n")
}

func routeLine(row []string, widths []int) string {
	cols := make([]string, 0, len(row))
	for i, cell := range row {
		cols = append(cols, fmt.Sprintf("%-*s", widths[i], cell))
	}
	return "  " + strings.TrimRight(strings.Join(cols, "  "), " ")
}

func columnWidths(cells [][]string) []int {
	widths := make([]int, len(cells[0]))
	for _, row := range cells {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	return widths
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
