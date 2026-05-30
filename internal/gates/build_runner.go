package gates

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/samuelnp/centinela/internal/config"
)

// targetFailure pairs a build target with the error it produced.
type targetFailure struct {
	Target config.BuildTarget
	Err    error
}

// buildTarget cross-compiles one target by running command for the given
// GOOS/GOARCH. The command is parsed into argv with strings.Fields and exec'd
// directly (no `sh -c`) to avoid shell injection. stdout is discarded; stderr
// is captured and, on failure, its first line is folded into the returned
// error alongside the target identifier.
func buildTarget(command string, t config.BuildTarget) error {
	argv := strings.Fields(command)
	if len(argv) == 0 {
		return fmt.Errorf("%s/%s: empty build command", t.GOOS, t.GOARCH)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = io.Discard
	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Env = buildEnv(t)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s/%s: %s", t.GOOS, t.GOARCH, firstStderrLine(stderr.String(), err))
	}
	return nil
}

// firstStderrLine returns the first non-empty stderr line, falling back to the
// exec error string when stderr is empty (e.g. command not found).
func firstStderrLine(stderr string, runErr error) string {
	for _, line := range strings.Split(stderr, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return runErr.Error()
}

// runTargets cross-compiles every target with a bounded worker pool sized to
// GOMAXPROCS, collecting per-target failures. Order of failures is not
// guaranteed; callers that need stability should sort.
func runTargets(command string, targets []config.BuildTarget) []targetFailure {
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan config.BuildTarget)
	var mu sync.Mutex
	var failures []targetFailure
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				if err := buildTarget(command, t); err != nil {
					mu.Lock()
					failures = append(failures, targetFailure{Target: t, Err: err})
					mu.Unlock()
				}
			}
		}()
	}
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	return failures
}
