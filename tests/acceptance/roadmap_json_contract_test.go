package acceptance_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const rmcBody = `{"phases":[{"name":"Backlog","features":[{"name":"deferred"}]},` +
	`{"name":"Q1","features":[{"name":"auth-service"},{"name":"billing-api"},` +
	`{"name":"checkout-ui","dependsOn":["auth-service"]},` +
	`{"name":"reporting","dependsOn":["billing-api"]}]}]}`

func rmcView(t *testing.T, dir string) (roadmap.RoadmapView, string) {
	t.Helper()
	out, _, code := rmcRun(t, dir, "roadmap", "--json")
	if code != 0 {
		t.Fatalf("roadmap --json exit=%d", code)
	}
	var v roadmap.RoadmapView
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	return v, out
}

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: roadmap --json emits ordered phases and features with counts
// Scenario: readiness is empty for a done feature; status alone carries the signal
// Scenario: readiness is empty for an in-progress feature; status alone carries the signal
// Scenario: readiness is "ready" for a planned feature whose dependencies are all done
// Scenario: readiness is "blocked" for a planned feature with an unmet dependency
// Scenario: status/readiness convention example row
// Scenario: Non-schedulable phases are excluded from roadmap --json
// Scenario: roadmap --json is byte-identical across two consecutive runs
func TestRoadmapViewJSONFullContract(t *testing.T) {
	d := rmcProject(t, rmcBody)
	rmcSeed(t, d, "auth-service", "done")
	rmcSeed(t, d, "billing-api", "code")
	v, out := rmcView(t, d)
	if len(v.Phases) != 1 || v.Phases[0].Name != "Q1" || len(v.Phases[0].Features) != 4 {
		t.Fatalf("Backlog excluded; Q1 has 4 ordered features: %+v", v.Phases)
	}
	if v.Counts != (roadmap.StatusCounts{Planned: 2, InProgress: 1, Done: 1}) {
		t.Fatalf("counts = %+v", v.Counts)
	}
	f := v.Phases[0].Features
	if f[0].Status != "done" || f[0].Readiness != "" || f[1].Status != "in-progress" || f[1].Readiness != "" {
		t.Fatalf("done/in-progress rows must omit readiness: %+v", f[:2])
	}
	if f[2].Readiness != "ready" || f[2].BlockedBy != nil {
		t.Fatalf("checkout-ui must be ready with no blockedBy: %+v", f[2])
	}
	if f[3].Readiness != "blocked" || len(f[3].BlockedBy) != 1 || f[3].BlockedBy[0] != "billing-api" {
		t.Fatalf("reporting must be blocked by billing-api: %+v", f[3])
	}
	if _, out2 := rmcView(t, d); out != out2 {
		t.Fatalf("roadmap --json must be byte-identical across runs")
	}
}

// Acceptance: specs/roadmap-json-contract.feature
// Scenario: roadmap ready --json emits the ready feature names in declared order
// Scenario: ready --json set is identical to the readiness:ready set in roadmap --json
// Scenario: roadmap ready --json is byte-identical across two consecutive runs
func TestRoadmapReadyJSONMatchesViewReadySet(t *testing.T) {
	d := rmcProject(t, rmcBody)
	rmcSeed(t, d, "auth-service", "done")
	rmcSeed(t, d, "billing-api", "code")
	out, _, code := rmcRun(t, d, "roadmap", "ready", "--json")
	if code != 0 {
		t.Fatalf("ready --json exit=%d", code)
	}
	var names []string
	if err := json.Unmarshal([]byte(out), &names); err != nil {
		t.Fatalf("invalid JSON array: %v\n%s", err, out)
	}
	if strings.Join(names, ",") != "checkout-ui" {
		t.Fatalf("ready set = %v want [checkout-ui]", names)
	}
	v, _ := rmcView(t, d)
	var ready []string
	for _, ft := range v.Phases[0].Features {
		if ft.Readiness == "ready" {
			ready = append(ready, ft.Name)
		}
	}
	if strings.Join(ready, ",") != strings.Join(names, ",") {
		t.Fatalf("ready --json %v != view ready set %v", names, ready)
	}
	if out2, _, _ := rmcRun(t, d, "roadmap", "ready", "--json"); out != out2 {
		t.Fatalf("ready --json must be byte-identical across runs")
	}
}
