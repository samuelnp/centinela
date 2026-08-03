package main

import (
	"strings"
	"testing"
)

// The skeleton is captured straight into prompt text, so a rejected invocation
// must write NOTHING to stdout: a partial or explanatory byte there becomes
// part of an agent's evidence file. Role parsing happens before any derivation
// precisely so this holds.
func TestEvidenceSchemaErrorPathsPrintNothingToStdout(t *testing.T) {
	chdirEvidenceTemp(t)
	out := captureStdout(t, func() {
		if err := runEvidenceSchema(nil, []string{"bogus"}); err == nil {
			t.Fatal("unknown role must fail")
		} else if !strings.Contains(err.Error(), "unknown role") {
			t.Fatalf("error must name the problem, got %v", err)
		}
	})
	if out != "" {
		t.Fatalf("unknown role wrote %d bytes to stdout: %q", len(out), out)
	}
}

// Arity is enforced by the command's Args validator, i.e. before RunE and
// before anything can be printed. Asserted through the command object so the
// guarantee survives a future switch to a variadic signature.
func TestEvidenceSchemaRejectsWrongArity(t *testing.T) {
	for _, args := range [][]string{{}, {"gatekeeper", "extra"}} {
		out := captureStdout(t, func() {
			if err := evidenceSchemaCmd.Args(evidenceSchemaCmd, args); err == nil {
				t.Fatalf("args %v must be rejected", args)
			}
		})
		if out != "" {
			t.Fatalf("args %v wrote to stdout: %q", args, out)
		}
	}
}
