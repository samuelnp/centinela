package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

const markers = "<<<<<<< HEAD\n{}\n=======\n{}\n>>>>>>> other\n"

// resolveStub answers ls-files/show/add from a table and records staging.
func resolveStub(unmerged map[string]string, blobs map[string]string, staged *[]string) roadmap.StageRunner {
	return func(_ string, args ...string) (string, error) {
		switch args[0] {
		case "ls-files":
			return unmerged[args[len(args)-1]], nil
		case "show":
			if body, ok := blobs[args[1]]; ok {
				return body, nil
			}
			return "", errors.New("no such stage")
		case "add":
			*staged = append(*staged, args[2:]...)
			return "", nil
		}
		return "", nil
	}
}

func resolveRepo(t *testing.T, onDisk string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow/roadmap.json", []byte(onDisk), 0o644); err != nil {
		t.Fatal(err)
	}
}

func doc(entries string) string {
	return `{"phases":[{"name":"P","features":[]},{"name":"Backlog","features":[` + entries + `]}]}`
}

func TestResolveCommandUnionsAndStagesBoth(t *testing.T) {
	resolveRepo(t, markers)
	var staged []string
	run := resolveStub(
		map[string]string{".workflow/roadmap.json": "unmerged\n", "ROADMAP.md": ""},
		map[string]string{
			":1:.workflow/roadmap.json": doc(""),
			":2:.workflow/roadmap.json": doc(`{"name":"ours","deferredAt":"2026-02-01T00:00:00Z"}`),
			":3:.workflow/roadmap.json": doc(`{"name":"theirs","deferredAt":"2026-01-01T00:00:00Z"}`),
		}, &staged)
	if err := resolveRoadmapState(".", run); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(".workflow/roadmap.json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "<<<<") {
		t.Fatalf("conflict markers survived:\n%s", body)
	}
	for _, slug := range []string{"ours", "theirs"} {
		if !strings.Contains(string(body), slug) {
			t.Fatalf("%s was dropped:\n%s", slug, body)
		}
	}
	if _, err := os.Stat("ROADMAP.md"); err != nil {
		t.Fatalf("ROADMAP.md must be regenerated: %v", err)
	}
	if strings.Join(staged, ",") != ".workflow/roadmap.json,ROADMAP.md" {
		t.Fatalf("staged = %v", staged)
	}
}
