package acceptance

// parseGoNonVerbose RECOGNIZES plain `go test` output and reports SkipData
// false: the shape is known, but it structurally carries no skip information.
// That is a quiet "no skip data", not a warning — see Summary.SkipData. Scope
// is irrelevant here precisely because there is nothing to attribute.
func parseGoNonVerbose(output string, _ Scope) (Summary, bool) {
	if !goPackageLine.MatchString(output) {
		return Summary{}, false
	}
	return Summary{Shape: ShapeGoNonVerbose}, true
}
