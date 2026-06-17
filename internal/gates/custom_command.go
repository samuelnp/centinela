package gates

import (
	"fmt"
	"strings"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

const (
	customBlobCap  = 4096 // byte cap for a blob-mode Details entry
	customLineCap  = 200  // max Details entries in lines mode
	customTruncMsg = "… (truncated)"
)

// customGates runs every enabled [[gates.custom]] entry and returns their
// Results. Disabled entries are skipped. Diff-aware entries receive the changed
// file set from filter; all others (and a nil filter) full-scan.
func customGates(cfg *config.Config, filter *gitdiff.Set) []Result {
	var results []Result
	for _, g := range cfg.Gates.CustomGates {
		if !g.Enabled {
			continue
		}
		var changed []string
		if g.DiffAware && filter != nil {
			changed = filter.Paths()
		}
		timeout := time.Duration(g.TimeoutSeconds) * time.Second
		output, code, timedOut := runCustom(g.Command, timeout, changed)
		results = append(results, customResult(g, output, code, timedOut))
	}
	return results
}

// customResult maps a command's outcome onto the shared Result contract:
// timeout => Fail, exit 0 => Pass, non-zero => Fail (severity fail) or Warn.
func customResult(g config.CustomGate, output string, code int, timedOut bool) Result {
	if timedOut {
		msg := fmt.Sprintf("%s timed out after %ds", g.Name, g.TimeoutSeconds)
		return Result{Name: g.Name, Status: Fail, Message: msg, Details: []string{msg}}
	}
	if code == 0 {
		return Result{Name: g.Name, Status: Pass, Message: g.Name + " passed"}
	}
	status := Fail
	if g.Severity == "warn" {
		status = Warn
	}
	return Result{
		Name:    g.Name,
		Status:  status,
		Message: fmt.Sprintf("%s failed (exit %d)", g.Name, code),
		Details: customDetails(g, output, code),
	}
}

// customDetails builds the Details slice per the gate's output mode, always
// emitting a generic fallback so a Fail is never empty.
func customDetails(g config.CustomGate, output string, code int) []string {
	if strings.TrimSpace(output) == "" {
		return []string{fmt.Sprintf("%s failed (exit %d) with no output", g.Name, code)}
	}
	if g.Output == "lines" {
		return lineDetails(output)
	}
	return blobDetails(output)
}

// blobDetails returns the whole output as one entry, byte-capped with a marker.
func blobDetails(output string) []string {
	if len(output) > customBlobCap {
		return []string{output[:customBlobCap] + customTruncMsg}
	}
	return []string{output}
}

// lineDetails returns one entry per non-empty line, bounded with an overflow
// marker so a command emitting thousands of lines stays readable and the audit
// ratchet fingerprints each violation individually.
func lineDetails(output string) []string {
	var lines []string
	for _, l := range strings.Split(output, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		lines = append(lines, l)
	}
	if len(lines) > customLineCap {
		extra := len(lines) - customLineCap
		lines = append(lines[:customLineCap], fmt.Sprintf("… (%d more)", extra))
	}
	return lines
}
