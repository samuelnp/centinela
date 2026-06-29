package main

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/analyze"
	"github.com/samuelnp/centinela/internal/ui"
)

// emitSetupDirective prints the PROJECT.md-missing directive. If the repo
// already has source (brownfield), it routes the agent to draft PROJECT.md
// from the codebase via analyze/synthesize instead of cold-questioning the
// user. A truly empty (greenfield) repo keeps the question-based setup path.
func emitSetupDirective() {
	if analyze.HasSource(".") {
		fmt.Println("CENTINELA DIRECTIVE: brownfield setup. Existing code detected — draft PROJECT.md from the codebase, then confirm. Do not interrogate the user.")
		fmt.Println(ui.RenderBrownfieldSetupNeeded())
		return
	}
	fmt.Println("CENTINELA DIRECTIVE: setup required. Ask setup questions and write PROJECT.md.")
	fmt.Println(ui.RenderSetupNeeded())
}
