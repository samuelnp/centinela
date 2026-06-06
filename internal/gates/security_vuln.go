package gates

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samuelnp/centinela/internal/config"
)

const vulnName = "G-Vuln: Dependency Audit"

// vulnKey identifies a finding by affected package and vulnerability ID, used
// to de-duplicate the same CVE reported by more than one tool.
type vulnKey struct{ Pkg, ID string }

// checkVuln runs every configured vuln scanner whole-project (ignoring any diff
// filter, like the import-graph gate), de-dups findings by (package, id), and
// folds the outcome into one Result. Any finding -> Warn (never blocks); none
// found across present tools -> Pass; all tools absent/nothing-to-scan -> Skip;
// a tool that ran but produced unusable output -> Warn.
func checkVuln(cfg *config.Config) Result {
	r := Result{Name: vulnName}
	findings := map[vulnKey]bool{}
	var present, warns []string
	for _, tool := range cfg.Gates.Security.Vuln.Tools {
		if !toolPresent(tool) {
			continue
		}
		present = append(present, tool)
		fs, note := runVulnTool(tool)
		for _, k := range fs {
			findings[k] = true
		}
		if note != "" {
			warns = append(warns, note)
		}
	}
	if len(present) == 0 {
		r.Status = Skip
		r.Message = "No vuln scanner installed; dependency audit skipped."
		return r
	}
	return foldVuln(r, findings, warns)
}

// foldVuln maps the aggregated findings and per-tool warnings to a Result.
// Findings dominate (Warn with sorted Details); otherwise a tool-level warning
// surfaces as Warn; a fully clean run is a Pass.
func foldVuln(r Result, findings map[vulnKey]bool, warns []string) Result {
	if len(findings) > 0 {
		r.Status = Warn
		r.Message = "Vulnerable dependencies found (warning — does not block validate):"
		r.Details = vulnDetails(findings)
		return r
	}
	if len(warns) > 0 {
		r.Status = Warn
		r.Message = "Dependency audit incomplete:"
		r.Details = warns
		return r
	}
	r.Status = Pass
	r.Message = "No known-vulnerable dependencies found."
	return r
}

// vulnDetails renders the de-duplicated finding set as sorted "pkg: id" lines.
func vulnDetails(findings map[vulnKey]bool) []string {
	lines := make([]string, 0, len(findings))
	for k := range findings {
		lines = append(lines, fmt.Sprintf("%s: %s", k.Pkg, k.ID))
	}
	sort.Strings(lines)
	return lines
}

// runVulnTool dispatches to the per-tool runner, returning its findings plus an
// optional non-fatal note (a parse/tool warning) for the named tool.
func runVulnTool(tool string) ([]vulnKey, string) {
	out, stderr, runErr := runScanner(tool, vulnArgs(tool)...)
	if runErr == errScanTimeout {
		return nil, tool + ": timed out"
	}
	keys, perr := parseVuln(tool, out)
	if perr != nil {
		return nil, fmt.Sprintf("%s: %s", tool, firstStderrLine(string(stderr), perr))
	}
	return keys, ""
}

// vulnArgs returns the JSON-output argv for a supported scanner.
func vulnArgs(tool string) []string {
	if tool == "osv-scanner" {
		return []string{"--format", "json", "-r", "."}
	}
	return strings.Fields("-json ./...")
}
