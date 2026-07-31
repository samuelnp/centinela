package acceptance

import "regexp"

// goVerboseResult matches `go test -v` per-test result lines, including the
// indented form subtests use.
var goVerboseResult = regexp.MustCompile(`^[ \t]*--- (PASS|FAIL|SKIP): `)

// goPackageLine matches the per-package result lines go test prints, capturing
// the package path. It is the block terminator scanTextReport attributes on.
var goPackageLine = regexp.MustCompile(`(?m)^(?:ok|FAIL|\?)[ \t]+(\S+)[ \t]`)
