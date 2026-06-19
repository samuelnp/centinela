package importgraph

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// execRunner is the default Runner backed by exec.Command. stdout is returned
// on success; on failure the first stderr line is folded into the error for an
// actionable diagnostic (mirrors internal/golist).
func execRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if line := firstStderrLine(stderr.String()); line != "" {
			return nil, fmt.Errorf("%s: %s", name, line)
		}
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return stdout.Bytes(), nil
}

// firstStderrLine returns the first non-empty trimmed line of stderr.
func firstStderrLine(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return ""
}

// onPath reports whether an executable is resolvable on PATH. Wrapped in a var
// so selection/backends can be unit-tested with a fake.
var onPath = func(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
