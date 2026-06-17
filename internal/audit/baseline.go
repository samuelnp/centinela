package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Baseline is the committed, deterministic ratchet snapshot. Gates are sorted by
// name and each gate's fingerprints by Hash, so re-recording an unchanged repo
// yields a byte-identical file and clean git diffs (AC-7).
type Baseline struct {
	Scheme  string      `json:"scheme"`  // == fingerprintScheme; stale if mismatched
	Version int         `json:"version"` // file-format version, currently 1
	Gates   []GateEntry `json:"gates"`   // sorted by Gate name
}

// GateEntry groups one gate's baselined fingerprints, sorted by Hash.
type GateEntry struct {
	Gate         string        `json:"gate"`         // Result.Name
	Fingerprints []Fingerprint `json:"fingerprints"` // sorted by Hash
}

// SchemeStale reports whether a loaded baseline was recorded under a different
// fingerprint scheme; callers surface it as "re-run audit baseline" non-blocking.
func (b Baseline) SchemeStale() bool {
	return b.Scheme != "" && b.Scheme != fingerprintScheme
}

// sortBaseline canonicalizes a baseline in place: gates by name, fingerprints by
// Hash. Determinism is enforced here so every write path goes through one order.
func sortBaseline(b *Baseline) {
	for i := range b.Gates {
		fps := b.Gates[i].Fingerprints
		sort.Slice(fps, func(a, c int) bool { return fps[a].Hash < fps[c].Hash })
	}
	sort.Slice(b.Gates, func(i, j int) bool { return b.Gates[i].Gate < b.Gates[j].Gate })
}

// Save writes the baseline as deterministic, sorted, indented JSON with a
// trailing newline.
func Save(path string, b Baseline) error {
	sortBaseline(&b)
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling baseline: %w", err)
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating baseline dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing baseline %s: %w", path, err)
	}
	return nil
}

// Load reads a baseline from disk. The bool reports whether the file existed; a
// missing file is not an error (the ratchet treats it as "nothing baselined").
func Load(path string) (Baseline, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Baseline{}, false, nil
		}
		return Baseline{}, false, fmt.Errorf("reading baseline %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, true, fmt.Errorf("parsing baseline %s: %w", path, err)
	}
	return b, true, nil
}
