package evidence

import (
	"fmt"
	"strconv"
	"strings"
)

// parseCoverage accepts a percentage figure with an optional trailing "%"
// (e.g. "85", "85.0", "85%") and returns it as a float in 0..100.
func parseCoverage(s string) (float64, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(s), "%")
	f, err := strconv.ParseFloat(strings.TrimSpace(trimmed), 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a coverage percentage", s)
	}
	if f < 0 || f > 100 {
		return 0, fmt.Errorf("coverage %v out of range (0..100)", f)
	}
	return f, nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	}
	return false, fmt.Errorf("cannot parse %q as bool", s)
}
