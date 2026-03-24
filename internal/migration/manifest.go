package migration

import (
	"sort"

	"github.com/samuelnp/centinela/internal/scaffold"
)

const CurrentDocVersion = "1"

func managedPaths() ([]string, error) {
	paths, err := scaffold.ListAssetFiles("docs/architecture/", ".md")
	if err != nil {
		return nil, err
	}
	paths = append(paths, "CLAUDE.md", "PROJECT.md.template")
	sort.Strings(paths)
	return paths, nil
}
