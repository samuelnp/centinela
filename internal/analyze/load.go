package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ErrNoInventory is returned by Load when the inventory file does not exist, so
// callers can surface a "run centinela analyze first" message distinct from a
// malformed or schema-drifted file.
var ErrNoInventory = errors.New("no analysis inventory")

// Load reads and decodes an Inventory written by Save. A missing file yields
// ErrNoInventory (wrapped); malformed JSON or a SchemaVersion mismatch yields a
// distinct, actionable error. The decoded Inventory is returned only when the
// schema matches the current contract.
func Load(path string) (Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Inventory{}, fmt.Errorf("%w at %s", ErrNoInventory, path)
		}
		return Inventory{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var inv Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		return Inventory{}, fmt.Errorf("malformed inventory %s: %w", path, err)
	}
	if inv.SchemaVersion != SchemaVersion {
		return Inventory{}, fmt.Errorf(
			"inventory %s has schema v%d, want v%d — re-run centinela analyze",
			path, inv.SchemaVersion, SchemaVersion)
	}
	return inv, nil
}
