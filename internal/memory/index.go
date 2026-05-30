package memory

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

// indexRecord is a recall-oriented projection of an Entry for index.json.
type indexRecord struct {
	ID             string    `json:"id"`
	Feature        string    `json:"feature"`
	Step           string    `json:"step"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Tags           []string  `json:"tags"`
	SourceArtifact string    `json:"sourceArtifact"`
	CreatedAt      time.Time `json:"createdAt"`
}

// regenerateIndex rebuilds index.json from the entry files, which remain the
// source of truth (D5). The index is a regenerable cache.
func regenerateIndex() error {
	entries := loadEntries()
	records := make([]indexRecord, 0, len(entries))
	for _, e := range entries {
		records = append(records, indexRecord{
			ID:             e.ID,
			Feature:        e.Feature,
			Step:           e.Step,
			Type:           e.Type,
			Title:          e.Title,
			Tags:           e.Tags,
			SourceArtifact: e.SourceArtifact,
			CreatedAt:      e.CreatedAt,
		})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	if err := os.MkdirAll(ledgerDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexFile, append(data, '\n'), 0o644)
}
