package orchestration

import (
	"strings"
	"testing"
)

func TestRoutingDirective_NamesUnroutedRolesFloorsAndStatics(t *testing.T) {
	line := RoutingDirective("f", []Role{RolePlanner}, nil, nil, RoleModels{})
	for _, want := range []string{
		"routing (dynamic): unrouted [planner]",
		"floors: planner>=balanced",
		"static: planner=reasoning",
		"decide: centinela route set f <role> <tier>",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("directive missing %q: %s", want, line)
		}
	}
}

func TestRoutingDirective_SilentOnceAllRolesRouted(t *testing.T) {
	routed := map[string]string{"planner": "reasoning"}
	if line := RoutingDirective("f", []Role{RolePlanner}, routed, nil, RoleModels{}); line != "" {
		t.Fatalf("a fully routed step must emit nothing, got %q", line)
	}
	if line := RoutingDirective("f", nil, nil, nil, RoleModels{}); line != "" {
		t.Fatalf("no roles must emit nothing, got %q", line)
	}
}

func TestRoutingDirective_FloorlessRolesOmitTheFloorsSegment(t *testing.T) {
	line := RoutingDirective("f", []Role{RoleSeniorEngineer}, nil, nil, RoleModels{})
	if strings.Contains(line, "floors:") {
		t.Fatalf("a floorless step must omit the floors segment: %s", line)
	}
	if !strings.Contains(line, "static: senior-engineer=reasoning") {
		t.Fatalf("static reference missing: %s", line)
	}
}

func TestRoutingDirective_MixedRolesAndConfiguredFloors(t *testing.T) {
	roles := []Role{RoleSeniorEngineer, RoleUXUISpecialist}
	models := RoleModels{"senior-engineer": {Tier: "fast"}}
	floors := map[string]string{"ux-ui-specialist": "balanced"}
	line := RoutingDirective("f", roles, map[string]string{"ux-ui-specialist": "fast"}, floors, models)
	if !strings.Contains(line, "unrouted [senior-engineer]") || strings.Contains(line, "ux-ui") {
		t.Fatalf("only un-routed roles belong on the line: %s", line)
	}
	if !strings.Contains(line, "static: senior-engineer=fast") {
		t.Fatalf("the static column must reflect config, got: %s", line)
	}
}
