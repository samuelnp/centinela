package roadmap

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseScores parses a CSV of exactly six integers in the order
// ac,uv,dc,dep,ee,overall, validating each is 1-10. There is no minimum for
// overall: the self-graded threshold gate was deleted in every profile, so the
// scores are RECORDED here, not enforced. All validation happens before any
// write at the call site.
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
	// Pass the richer message through: the CSV parse above makes a structural
	// fault impossible here, so the range error is the only one that can reach
	// this point and naming the field + value is a strict improvement.
	if err := validateScoreRange(s); err != nil {
		return QualityScores{}, err
	}
	return s, nil
}
