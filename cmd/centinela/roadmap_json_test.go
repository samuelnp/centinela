package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const q1Roadmap = `{"phases":[{"name":"Q1","features":[{"name":"auth-service"},` +
	`{"name":"billing-api"},{"name":"checkout-ui","dependsOn":["auth-service"]},` +
	`{"name":"reporting","dependsOn":["billing-api"]}]}]}`

// roadmap --json emits a valid RoadmapView with ordered phases and counts.
func TestRunRoadmap_JSON(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json", q1Roadmap)
	markCmdDone(t, "auth-service")
	seedWF(t, "billing-api", "code")
	roadmapViewJSON = true
	defer func() { roadmapViewJSON = false }()
	out := captureStdout(t, func() {
		if err := runRoadmap(nil, nil); err != nil {
			t.Fatalf("runRoadmap: %v", err)
		}
	})
	var v roadmap.RoadmapView
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(v.Phases) != 1 || len(v.Phases[0].Features) != 4 {
		t.Fatalf("want Q1 with 4 features, got %+v", v.Phases)
	}
	if v.Counts != (roadmap.StatusCounts{Planned: 2, InProgress: 1, Done: 1}) {
		t.Fatalf("counts = %+v", v.Counts)
	}
}

// ready --json emits the ready names in declared order and equals the
// readiness:"ready" set from roadmap --json.
func TestRunRoadmapReady_JSON_MatchesViewReadySet(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json", q1Roadmap)
	markCmdDone(t, "auth-service")
	seedWF(t, "billing-api", "code")
	readyJSON = true
	defer func() { readyJSON = false }()
	out := captureStdout(t, func() {
		if err := runRoadmapReady(nil, nil); err != nil {
			t.Fatalf("runRoadmapReady: %v", err)
		}
	})
	var names []string
	if err := json.Unmarshal([]byte(out), &names); err != nil {
		t.Fatalf("invalid JSON array: %v\n%s", err, out)
	}
	if strings.Join(names, ",") != "checkout-ui" {
		t.Fatalf("ready set = %v want [checkout-ui]", names)
	}
	var ready []string
	for _, f := range roadmap.BuildView(loadOrFail(t)).Phases[0].Features {
		if f.Readiness == "ready" {
			ready = append(ready, f.Name)
		}
	}
	if strings.Join(ready, ",") != strings.Join(names, ",") {
		t.Fatalf("ready --json %v != view ready set %v", names, ready)
	}
}

func loadOrFail(t *testing.T) *roadmap.Roadmap {
	t.Helper()
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return r
}
