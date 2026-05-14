package worktree

import (
	"encoding/json"
	"os"
)

// patchTsconfigExclude adds entry to the "exclude" array of tsconfig.json.
// Tolerant of missing files (returns false, nil). Preserves all other keys.
// Idempotent.
func patchTsconfigExclude(path, entry string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		// Tolerate non-JSON or commented tsconfigs: do not crash the wizard.
		return false, nil
	}
	excludes := decodeExcludes(doc["exclude"])
	if hasExclude(excludes, entry) {
		return false, nil
	}
	excludes = append(excludes, entry)
	encoded, err := json.Marshal(excludes)
	if err != nil {
		return false, err
	}
	doc["exclude"] = encoded
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0644); err != nil {
		return false, err
	}
	return true, nil
}

func decodeExcludes(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil
	}
	return arr
}

func hasExclude(list []string, entry string) bool {
	for _, e := range list {
		if e == entry {
			return true
		}
	}
	return false
}
