package docgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadData(title string) (*Data, error) {
	if err := ValidateInputs(); err != nil {
		return nil, err
	}
	d := &Data{Title: title}
	d.Project = readFile("PROJECT.md")
	d.RoadmapText = readFile("ROADMAP.md")
	d.FeatureDocs = listFiles("docs/features/*.md")
	d.PlanDocs = listFiles("docs/plans/*.md")
	d.Specs, d.Scenarios = loadSpecs()
	d.RoadmapNodes = loadRoadmapNodes()
	d.Evidence = loadEvidence()
	d.States = loadStates()
	kb, err := loadKBPages()
	if err != nil {
		return nil, err
	}
	d.KB = kb
	return d, nil
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func listFiles(pattern string) []string {
	files, _ := filepath.Glob(pattern)
	sort.Strings(files)
	return files
}

func loadSpecs() ([]string, int) {
	s := listFiles("specs/*.feature")
	total := 0
	for _, p := range s {
		total += strings.Count(readFile(p), "Scenario:")
	}
	return s, total
}

func loadRoadmapNodes() []RoadmapNode {
	raw := readFile(".workflow/roadmap.json")
	var rm struct {
		Phases []struct {
			Features []struct {
				Name      string   `json:"name"`
				DependsOn []string `json:"dependsOn"`
			} `json:"features"`
		} `json:"phases"`
	}
	json.Unmarshal([]byte(raw), &rm) //nolint:errcheck
	out := []RoadmapNode{}
	for _, p := range rm.Phases {
		for _, f := range p.Features {
			out = append(out, RoadmapNode{
				Name:      f.Name,
				DependsOn: f.DependsOn,
			})
		}
	}
	return out
}
