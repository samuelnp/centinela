package acceptance

// add sums one summary into another. Used only ACROSS package blocks of the
// same shape, where the counts describe different packages' work.
func (s *Summary) add(o Summary) {
	s.Scenarios += o.Scenarios
	s.Passed += o.Passed
	s.Skipped += o.Skipped
	s.Pending += o.Pending
	s.Undefined += o.Undefined
}

// atLeast folds one shape's counts into another by per-field MAXIMUM. Used
// across DIFFERENT shapes, which routinely describe the same work twice — a
// godog run of 3 scenarios driven by one Go wrapper test is both "3 scenarios"
// and "1 Go test" — so summing would inflate the numbers the operator is told
// to act on. The boolean facts are unions: either being true is true of the run.
func (s *Summary) atLeast(o Summary) {
	s.Scenarios = max(s.Scenarios, o.Scenarios)
	s.Passed = max(s.Passed, o.Passed)
	s.Skipped = max(s.Skipped, o.Skipped)
	s.Pending = max(s.Pending, o.Pending)
	s.Undefined = max(s.Undefined, o.Undefined)
	s.GherkinZero = s.GherkinZero || o.GherkinZero
	s.Attributed = s.Attributed || o.Attributed
}
