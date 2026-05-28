package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, _ := os.Pipe()
	orig := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = orig })
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	fn()
	_ = w.Close()
	<-done
	return buf.String()
}

func TestEvidenceValidateSucceedsWhenNoFiles(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := runEvidenceValidate(nil, []string{"alpha"}); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
}

func TestEvidenceValidateEmitsFixHint(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	var validateErr error
	stderr := captureStderr(t, func() {
		validateErr = runEvidenceValidate(nil, []string{"alpha"})
	})
	if validateErr == nil {
		t.Fatal("expected non-nil error for missing inputs")
	}
	if !strings.Contains(stderr, "centinela evidence append") {
		t.Fatalf("expected fix hint in stderr, got %q", stderr)
	}
}
