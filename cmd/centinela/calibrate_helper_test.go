package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureCalibrate runs runCalibrate in dir with the given --json flag,
// returning stdout and any error. Restores stdout and the flag afterward.
func captureCalibrate(t *testing.T, dir string, asJSON bool) (string, error) {
	t.Helper()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	os.Chdir(dir)                     //nolint:errcheck

	oldFlag := calibrateJSON
	calibrateJSON = asJSON
	t.Cleanup(func() { calibrateJSON = oldFlag })

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runCalibrate(nil, nil)
	w.Close() //nolint:errcheck
	os.Stdout = oldStdout
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	return string(buf[:n]), err
}

// seedCalLog writes events.jsonl lines into dir/.workflow/telemetry.
func seedCalLog(t *testing.T, dir string, lines ...string) {
	t.Helper()
	td := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(td, "events.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
