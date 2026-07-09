package main

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// newEditCmd builds a fresh edit command for independent Changed() sentinel state.
func newEditCmd() *cobra.Command {
	editName, editDescription, editArchetype, editDependsOn = "", "", "", nil
	c := &cobra.Command{Use: "edit", RunE: runRoadmapEdit}
	c.Flags().StringVar(&editName, "name", "", "")
	c.Flags().StringVar(&editDescription, "description", "", "")
	c.Flags().StringVar(&editArchetype, "archetype", "", "")
	c.Flags().StringSliceVar(&editDependsOn, "depends-on", nil, "")
	return c
}

const editDepsRoadmap = `{"phases":[{"name":"Phase 1","features":[` +
	`{"name":"a"},{"name":"b","dependsOn":["a"]}]}]}`

// TestRunRoadmapEdit_Description edits one field through the command.
func TestRunRoadmapEdit_Description(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	c := newEditCmd()
	_ = c.Flags().Set("description", "Updated")
	if err := runRoadmapEdit(c, []string{"auth-service"}); err != nil {
		t.Fatalf("runRoadmapEdit: %v", err)
	}
	for _, p := range mustLoad(t).Phases {
		for _, f := range p.Features {
			if f.Name == "auth-service" && f.Description != "Updated" {
				t.Fatalf("description not applied: %+v", f)
			}
		}
	}
}

// TestRunRoadmapEdit_ClearDepsSentinel: an explicit empty --depends-on clears deps.
func TestRunRoadmapEdit_ClearDepsSentinel(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, editDepsRoadmap)
	c := newEditCmd()
	_ = c.Flags().Set("depends-on", "")
	if err := runRoadmapEdit(c, []string{"b"}); err != nil {
		t.Fatalf("runRoadmapEdit: %v", err)
	}
	if d := featB(t).DependsOn; len(d) != 0 {
		t.Fatalf("explicit empty --depends-on must clear: %v", d)
	}
}

// TestRunRoadmapEdit_UnchangedDeps: omitting --depends-on preserves deps.
func TestRunRoadmapEdit_UnchangedDeps(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, editDepsRoadmap)
	c := newEditCmd()
	_ = c.Flags().Set("description", "x")
	if err := runRoadmapEdit(c, []string{"b"}); err != nil {
		t.Fatalf("runRoadmapEdit: %v", err)
	}
	if d := featB(t).DependsOn; len(d) != 1 || d[0] != "a" {
		t.Fatalf("omitted --depends-on must preserve: %v", d)
	}
}

// TestRunRoadmapEdit_Error surfaces a package rejection (unknown slug).
func TestRunRoadmapEdit_Error(t *testing.T) {
	chdirIntoTemp(t)
	writeFile(t, roadmap.RoadmapFile, addRoadmap)
	if err := runRoadmapEdit(newEditCmd(), []string{"ghost"}); err == nil {
		t.Fatal("unknown slug must error")
	}
}

func mustLoad(t *testing.T) *roadmap.Roadmap {
	t.Helper()
	r, err := roadmap.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return r
}

func featB(t *testing.T) roadmap.Feature {
	t.Helper()
	for _, p := range mustLoad(t).Phases {
		for _, f := range p.Features {
			if f.Name == "b" {
				return f
			}
		}
	}
	t.Fatal("feature b missing")
	return roadmap.Feature{}
}
