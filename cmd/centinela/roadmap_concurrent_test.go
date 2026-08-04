package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// helperEnv makes this test binary re-exec itself as a WORKER that performs one
// real `roadmap defer`. Separate OS processes are the tier the verifier caught
// F1 at — in-process goroutines cannot exercise git's own index.lock, and it
// was a failed `git add` under that lock that made the loss silent.
const helperEnv = "CENTINELA_TEST_DEFER_SLUG"

// TestHelperConcurrentDefer is not a test: it is the worker process. It exits
// immediately unless the environment marks it as such.
func TestHelperConcurrentDefer(t *testing.T) {
	slug := os.Getenv(helperEnv)
	if slug == "" {
		t.Skip("worker entry point; driven by TestConcurrentDefersFromTwoProcesses")
	}
	deferSummary, deferSource = "concurrent finding", "race/eng"
	if err := runRoadmapDefer(nil, []string{slug}); err != nil {
		t.Fatalf("worker defer %s: %v", slug, err)
	}
}

// F1 regression: two `roadmap defer` PROCESSES racing in one checkout must both
// land. Before the roadmap-state lock this dropped one record 4 times in 5, and
// the winner's auto-commit left a clean tree so `git status` showed nothing.
func TestConcurrentDefersFromTwoProcesses(t *testing.T) {
	dir := syncGitRepo(t)
	slugs := []string{"race-one", "race-two", "race-three"}
	var wg sync.WaitGroup
	outputs := make([]string, len(slugs))
	for i, slug := range slugs {
		wg.Add(1)
		go func(i int, slug string) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=TestHelperConcurrentDefer")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), helperEnv+"="+slug)
			out, err := cmd.CombinedOutput()
			outputs[i] = string(out)
			if err != nil {
				t.Errorf("worker %s: %v\n%s", slug, err, out)
			}
		}(i, slug)
	}
	wg.Wait()
	body, err := os.ReadFile(dir + "/.workflow/roadmap.json")
	if err != nil {
		t.Fatal(err)
	}
	for i, slug := range slugs {
		if !strings.Contains(string(body), `"`+slug+`"`) {
			t.Fatalf("%q was destroyed by a concurrent mutation\nworker output:\n%s\nroadmap.json:\n%s",
				slug, outputs[i], body)
		}
	}
	// No worker may claim the record is in the tree when it is not; the
	// unverified wording is the only honest report of that state.
	for i, out := range outputs {
		if strings.Contains(out, "in your working tree") && !strings.Contains(string(body), slugs[i]) {
			t.Fatalf("worker %d claimed the tree holds a record that is gone:\n%s", i, out)
		}
	}
}
