package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestMainErrorPath(t *testing.T) {
	oldExec, oldErr, oldExit := executeRoot, stderrWriter, exitMain
	defer func() { executeRoot, stderrWriter, exitMain = oldExec, oldErr, oldExit }()

	var out bytes.Buffer
	code := 0
	executeRoot = func() error { return errors.New("boom") }
	stderrWriter = &out
	exitMain = func(c int) { code = c }
	main()
	if code != 1 || out.Len() == 0 {
		t.Fatalf("expected exit 1 and stderr output, code=%d out=%q", code, out.String())
	}
}

func TestMainSuccessPath(t *testing.T) {
	oldExec, oldExit := executeRoot, exitMain
	defer func() { executeRoot, exitMain = oldExec, oldExit }()

	called := false
	executeRoot = func() error { return nil }
	exitMain = func(int) { called = true }
	main()
	if called {
		t.Fatal("exit should not be called on success")
	}
}

func TestMainDefaultExecutePath(t *testing.T) {
	oldExec, oldExit := executeRoot, exitMain
	defer func() { executeRoot, exitMain = oldExec, oldExit }()
	executeRoot = func() error {
		rootCmd.SetArgs([]string{"--help"})
		return rootCmd.Execute()
	}
	exitMain = func(int) { t.Fatal("exit should not be called") }
	main()
}
