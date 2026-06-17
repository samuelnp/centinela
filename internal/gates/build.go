package gates

import (
	"fmt"
	"os"
	"sort"

	"github.com/samuelnp/centinela/internal/config"
)

// buildEnv returns the process environment with GOOS/GOARCH set to the target
// and CGO_ENABLED=0 — no C toolchain is needed for a pure-Go cross-compile,
// and disabling cgo keeps the builds deterministic and cache-friendly.
func buildEnv(t config.BuildTarget) []string {
	return append(os.Environ(),
		"GOOS="+t.GOOS,
		"GOARCH="+t.GOARCH,
		"CGO_ENABLED=0",
	)
}

// checkBuild cross-compiles every configured release target and folds the
// outcome into a single G-Build Result. Any target that fails to compile fails
// the gate, with one Details entry per broken target (sorted for stable output)
// naming its GOOS/GOARCH and the first compiler error line.
func checkBuild(cfg *config.Config) Result {
	b := cfg.Gates.Build
	r := Result{Name: "G-Build: Cross-Compile"}
	if len(b.Targets) == 0 {
		r.Status = Skip
		r.Message = "Build gate enabled but no targets configured."
		return r
	}
	failures := runTargets(b.Command, b.Targets)
	if len(failures) == 0 {
		r.Status = Pass
		r.Message = fmt.Sprintf("All %d release targets compile.", len(b.Targets))
		return r
	}
	r.Status = Fail
	r.Message = "These release targets failed to build:"
	for _, f := range failures {
		r.Details = append(r.Details, f.Err.Error())
	}
	sort.Strings(r.Details)
	return r
}
