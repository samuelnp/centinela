package main

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// ready --json with nothing ready emits an empty array, never null.
func TestRunRoadmapReady_JSON_EmptyArray(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Q1","features":[{"name":"a"},{"name":"b","dependsOn":["a"]}]}]}`)
	seedWF(t, "a", "code") // a in-progress → b blocked, nothing ready
	readyJSON = true
	defer func() { readyJSON = false }()
	out := captureStdout(t, func() {
		if err := runRoadmapReady(nil, nil); err != nil {
			t.Fatalf("runRoadmapReady: %v", err)
		}
	})
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("empty ready must be [] not null, got %q", out)
	}
}

// show --json dumps the persisted Roadmap verbatim: Backlog is retained and no
// derived status/readiness fields appear. list is a byte-identical alias.
func TestRunRoadmapShow_JSON_Verbatim(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Backlog","features":[{"name":"bl"}]},`+
			`{"name":"Q1","features":[{"name":"q"}]}]}`)
	roadmapShowJSON = true
	defer func() { roadmapShowJSON = false }()
	out := captureStdout(t, func() {
		if err := runRoadmapShow(nil, nil); err != nil {
			t.Fatalf("runRoadmapShow: %v", err)
		}
	})
	if !strings.Contains(out, `"Backlog"`) {
		t.Fatalf("show must dump persisted roadmap verbatim incl Backlog:\n%s", out)
	}
	if strings.Contains(out, `"status"`) || strings.Contains(out, `"readiness"`) {
		t.Fatalf("show must not carry derived fields:\n%s", out)
	}
}

// roadmap show (no flag) renders text identical to plain roadmap, with no JSON.
func TestRunRoadmapShow_Text_MatchesRoadmap(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, ".workflow/roadmap.json",
		`{"phases":[{"name":"Q1","features":[{"name":"q"}]}]}`)
	show := captureStdout(t, func() {
		if err := runRoadmapShow(nil, nil); err != nil {
			t.Fatalf("runRoadmapShow text: %v", err)
		}
	})
	plain := captureStdout(t, func() {
		if err := runRoadmap(nil, nil); err != nil {
			t.Fatalf("runRoadmap text: %v", err)
		}
	})
	if show != plain {
		t.Fatalf("show text must equal roadmap text:\n%q\n%q", show, plain)
	}
	if strings.Contains(show, `"phases"`) {
		t.Fatalf("text mode must not emit JSON:\n%s", show)
	}
}

// Every --json surface fails on a missing roadmap.json with no partial stdout.
func TestRoadmapJSON_MissingFile_NoPartialJSON(t *testing.T) {
	cases := []struct {
		name string
		flag *bool
		run  func() error
	}{
		{"roadmap", &roadmapViewJSON, func() error { return runRoadmap(nil, nil) }},
		{"ready", &readyJSON, func() error { return runRoadmapReady(nil, nil) }},
		{"show", &roadmapShowJSON, func() error { return runRoadmapShow(nil, nil) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			chdirIntoTemp(t)
			*c.flag = true
			defer func() { *c.flag = false }()
			var err error
			out := captureStdout(t, func() { err = c.run() })
			if err == nil || !strings.Contains(err.Error(), roadmap.RoadmapFile) {
				t.Fatalf("%s --json must error naming %s, got %v", c.name, roadmap.RoadmapFile, err)
			}
			if strings.TrimSpace(out) != "" {
				t.Fatalf("%s --json must emit no partial stdout JSON, got %q", c.name, out)
			}
		})
	}
}
