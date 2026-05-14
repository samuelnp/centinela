package worktree

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// appendIgnoreLine writes line to path if missing. Creates the file when it
// does not exist. Returns whether the file was changed.
func appendIgnoreLine(path, line string) (bool, error) {
	created, err := ensureFile(path)
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if containsLine(data, line) {
		return created, nil
	}
	buf := bytes.NewBuffer(data)
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		buf.WriteByte('\n')
	}
	buf.WriteString(line)
	buf.WriteByte('\n')
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return false, err
	}
	return true, nil
}

// containsLine reports whether the file's lines contain an exact match.
func containsLine(data []byte, line string) bool {
	target := strings.TrimSpace(line)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == target {
			return true
		}
	}
	return false
}
