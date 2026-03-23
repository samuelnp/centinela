package main

import (
	"os"
	"testing"
)

func TestRunHookSetupBranches(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	withStdin(t, "{}", func() {
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	os.WriteFile("PROJECT.md.template", []byte("x"), 0644) //nolint:errcheck
	withStdin(t, "{}", func() {
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	os.WriteFile("PROJECT.md", []byte("x"), 0644) //nolint:errcheck
	withStdin(t, "{}", func() {
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	os.WriteFile("ROADMAP.md", []byte("x"), 0644) //nolint:errcheck
	os.MkdirAll("docs/architecture", 0755)        //nolint:errcheck
	withStdin(t, "{}", func() {
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
	os.WriteFile("docs/architecture/production-readiness-prompt.md", []byte("x"), 0644) //nolint:errcheck
	withStdin(t, "{}", func() {
		if err := runHookSetup(nil, nil); err != nil {
			t.Fatal(err)
		}
	})
}
