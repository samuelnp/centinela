package main

import (
	"encoding/json"
	"fmt"

	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gates"
)

// docsLintJSONReport is the machine-readable shape of `centinela docs lint`.
type docsLintJSONReport struct {
	Scope        string   `json:"scope"`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	Files        int      `json:"files"`
	Inspected    int      `json:"inspected"`
	Undocumented int      `json:"undocumented"`
	Exempt       int      `json:"exempt"`
	Unparseable  int      `json:"unparseable"`
	Details      []string `json:"details"`
}

// docsLintStatusNames maps a gate status onto its JSON token.
var docsLintStatusNames = map[gates.Status]string{
	gates.Pass: "pass", gates.Fail: "fail", gates.Warn: "warn", gates.Skip: "skip",
}

// docsLintPayload renders the gate verdict for the changed-file scope.
func docsLintPayload(scope string, r gates.Result, rep docstring.Report) docsLintJSONReport {
	return docsLintJSONReport{
		Scope:        scope,
		Status:       docsLintStatusNames[r.Status],
		Message:      r.Message,
		Files:        rep.Files,
		Inspected:    rep.Inspected,
		Undocumented: len(rep.Violations),
		Exempt:       len(rep.Exemptions),
		Unparseable:  len(rep.ParseErrors),
		Details:      append([]string{}, r.Details...),
	}
}

// docsLintFullPayload renders the whole-repo backlog. Its status is "report",
// never a verdict — a full scan measures, it does not gate.
func docsLintFullPayload(rep docstring.Report) docsLintJSONReport {
	return docsLintJSONReport{
		Scope:        "full",
		Status:       "report",
		Message:      "whole-repo docstring backlog (report only)",
		Files:        rep.Files,
		Inspected:    rep.Inspected,
		Undocumented: len(rep.Violations),
		Exempt:       len(rep.Exemptions),
		Unparseable:  len(rep.ParseErrors),
		Details:      append(rep.Lines(), rep.ExemptionLines()...),
	}
}

// emitDocsLintJSON prints an indented JSON report on stdout.
func emitDocsLintJSON(r docsLintJSONReport) error {
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
