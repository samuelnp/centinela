package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/docstring"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/ui"
)

// runDocsLintFull reports the whole-repo documentation backlog. It is a
// measurement surface for adoption, deliberately never a gate: the gate is
// ratcheted to changed files precisely so a legacy backlog cannot make it
// permanently red, and this report must not smuggle that back in. Exit is
// always 0 — including when the scan finds thousands of gaps.
func runDocsLintFull(cfg *config.Config) error {
	opts := gates.DocstringOptions(cfg)
	files, err := docstring.Files(opts)
	if err != nil {
		return err
	}
	scanner, ok := docstring.For(docstring.GoLang)
	if !ok {
		return fmt.Errorf("no docstring scanner registered for %s", docstring.GoLang)
	}
	rep, err := scanner.Scan(files, opts)
	if err != nil {
		return err
	}
	if docsLintJSON {
		return emitDocsLintJSON(docsLintFullPayload(rep))
	}
	fmt.Println(ui.StyleBold.Render("Docstring backlog (full scan — report only)"))
	for _, line := range rep.Lines() {
		fmt.Println(ui.StyleMuted.Render("  · " + line))
	}
	for _, line := range rep.ExemptionLines() {
		fmt.Println(ui.StyleMuted.Render("  · " + line))
	}
	fmt.Printf("%d undocumented of %d exported identifiers across %d files (%d exempt)\n",
		len(rep.Violations), rep.Inspected, rep.Files, len(rep.Exemptions))
	return nil
}
