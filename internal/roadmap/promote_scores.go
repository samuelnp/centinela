package roadmap

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseScores parses a CSV of exactly six integers in the order
// ac,uv,dc,dep,ee,overall, validating each is 1-10 and overall >= the quality
// threshold. All validation happens before any write at the call site.
func ParseScores(csv string) (QualityScores, error) {
	parts := strings.Split(strings.TrimSpace(csv), ",")
	if len(parts) != 6 {
		return QualityScores{}, fmt.Errorf("--scores requires exactly six comma-separated integers (ac,uv,dc,dep,ee,overall)")
	}
	nums := make([]int, 6)
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return QualityScores{}, fmt.Errorf("--scores requires exactly six comma-separated integers (ac,uv,dc,dep,ee,overall)")
		}
		nums[i] = v
	}
	s := QualityScores{
		AcceptanceCriteria: nums[0], UserValue: nums[1], DefinitionClarity: nums[2],
		Dependencies: nums[3], EffortEstimation: nums[4], Overall: nums[5],
	}
	if err := validateScoreRange(s); err != nil {
		return QualityScores{}, fmt.Errorf("each score must be between 1 and 10")
	}
	if s.Overall < qualityThreshold {
		return QualityScores{}, fmt.Errorf("overall score must be at least %d", qualityThreshold)
	}
	return s, nil
}
