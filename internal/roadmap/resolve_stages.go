package roadmap

import (
	"fmt"
	"strings"
)

// StageRunner runs a git subcommand in dir and returns its stdout. It is the
// single seam resolve reads the conflict through, so the merge logic can be
// tested without staging a real conflict.
type StageRunner func(dir string, args ...string) (string, error)

// Stages holds the three versions git keeps for a conflicted path: :1: the
// merge base, :2: ours (the current branch), :3: theirs (the branch merging
// in). A side deleted on one branch has no stage and comes back empty.
type Stages struct {
	Base   []byte
	Ours   []byte
	Theirs []byte
}

// ReadStages loads the three merge stages of path from git's index. A path
// with no stages at all is not conflicted and returns ok=false, which the
// caller reports as "nothing to resolve" rather than as an error.
func ReadStages(dir, path string, run StageRunner) (Stages, bool, error) {
	if !Conflicted(dir, path, run) {
		return Stages{}, false, nil
	}
	var st Stages
	for _, spec := range []struct {
		stage string
		dst   *[]byte
	}{{"1", &st.Base}, {"2", &st.Ours}, {"3", &st.Theirs}} {
		out, err := run(dir, "show", ":"+spec.stage+":"+path)
		if err != nil {
			continue // a stage is absent when that side deleted the file
		}
		*spec.dst = []byte(out)
	}
	if len(st.Ours) == 0 && len(st.Theirs) == 0 {
		return Stages{}, false, fmt.Errorf("%s is conflicted but neither side has content", path)
	}
	return st, true, nil
}

// Conflicted reports whether git's index holds unmerged stages for path.
func Conflicted(dir, path string, run StageRunner) bool {
	out, err := run(dir, "ls-files", "--unmerged", "--", path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}
