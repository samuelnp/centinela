package main

import (
	"fmt"
	"os"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func roadmapStartError(err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("greenfield project requires %s before start\nWrite ROADMAP.md and %s, then run centinela roadmap validate", roadmap.RoadmapFile, roadmap.RoadmapFile)
	}
	return fmt.Errorf("greenfield project requires valid %s before start: %v\nUpdate %s to match ROADMAP.md, then run centinela roadmap validate", roadmap.RoadmapFile, err, roadmap.RoadmapFile)
}

func roadmapCommandError(err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("roadmap setup incomplete: missing %s\nWrite ROADMAP.md and %s, then run centinela roadmap validate", roadmap.RoadmapFile, roadmap.RoadmapFile)
	}
	return fmt.Errorf("roadmap setup incomplete: invalid %s: %v\nUpdate %s to match ROADMAP.md, then run centinela roadmap validate", roadmap.RoadmapFile, err, roadmap.RoadmapFile)
}
