package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestRunHookPrewriteNonBlockingPaths(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	os.WriteFile("centinela.toml", []byte("[workflow]\ncode_dirs=[\"/src/\"]\n"), 0644)                                                         //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                              //nolint:errcheck
	workflow.Save(&workflow.Workflow{Feature: "f", CurrentStep: "code", Steps: map[string]workflow.StepState{"code": {Status: "in-progress"}}}) //nolint:errcheck

	cases := []string{
		"{}",
		"not-json",
		`{"tool_input":{}}`,
		`{"tool_input":{"filePath":"/tmp/outside.go"}}`,
		`{"tool_input":{"filePath":"` + d + `/README.md"}}`,
		`{"tool_input":{"filePath":"` + d + `/docs/features/f.md"}}`,
		`{"tool_input":{"filePath":"` + d + `/src/a.go"}}`,
	}
	for _, in := range cases {
		in := in
		withStdin(t, in, func() {
			if err := runHookPrewrite(nil, nil); err != nil {
				t.Fatalf("runHookPrewrite(%q): %v", in, err)
			}
		})
	}
}
