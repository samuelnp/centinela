package acceptance

import "strings"

// textBlock is one package's accumulated results. `go test -v ./...` buffers
// per package, so every package's lines are CONTIGUOUS and terminated by that
// package's own ok/FAIL line — which is what makes attribution exact rather
// than a guess, for the Gherkin summary as much as for the Go result lines.
type textBlock struct {
	gherkin    Summary
	sawGherkin bool
	gotest     Summary
}

// scanTextReport walks the output ONCE, binding both text shapes to the package
// block that produced them and folding a block in only when the tier
// attribution accepts its package.
//
// sawGherkin / sawGo report whether the SHAPE was present anywhere, including
// in blocks that attribution discarded. That distinction matters: a whole-repo
// `-v` run whose skips all belong to other tiers still carries skip data, and
// must not fall through to the "carries no skip data" note.
func scanTextReport(output string, scope Scope) (gherkin, gotest Summary, sawGherkin, sawGo bool) {
	gherkin = Summary{Shape: ShapeCucumber, SkipData: true, Attributed: scope == ScopeMixed}
	gotest = Summary{Shape: ShapeGoVerbose, SkipData: true, Attributed: scope == ScopeMixed}
	var block textBlock
	flush := func(pkg string, attributable bool) {
		if attributable && scope.counts(pkg) {
			if block.sawGherkin {
				gherkin.add(block.gherkin)
				gherkin.GherkinZero = gherkin.GherkinZero || block.gherkin.GherkinZero
			}
			gotest.add(block.gotest)
		}
		block = textBlock{}
	}
	for _, line := range strings.Split(output, "\n") {
		switch {
		case addGherkinLine(&block, line):
			sawGherkin = true
		case addGoLine(&block, line):
			sawGo = true
		default:
			if m := goPackageLine.FindStringSubmatch(line); m != nil {
				flush(m[1], true)
			}
		}
	}
	// A trailing block with no terminator is only trustworthy when the command
	// itself was acceptance-scoped; under ScopeMixed it is unattributable.
	flush("", scope != ScopeMixed)
	return gherkin, gotest, sawGherkin, sawGo
}

// addGherkinLine records a Gherkin summary against the open block. The LAST
// summary in a block wins: a runner that prints per-feature summaries ends with
// the aggregate one.
func addGherkinLine(block *textBlock, line string) bool {
	s, ok := cucumberSummaryOf(line)
	if !ok {
		return false
	}
	block.gherkin, block.sawGherkin = s, true
	return true
}

// addGoLine records one `go test -v` per-test result against the open block.
func addGoLine(block *textBlock, line string) bool {
	m := goVerboseResult.FindStringSubmatch(line)
	if m == nil {
		return false
	}
	addGoResult(&block.gotest, m[1])
	return true
}
