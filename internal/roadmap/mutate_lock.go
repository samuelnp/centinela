package roadmap

import (
	"fmt"

	"github.com/samuelnp/centinela/internal/filelock"
)

// lockRoadmapState serializes the whole read → modify → write of roadmap.json
// against other processes mutating the same checkout, and returns a release the
// caller MUST defer.
//
// Without it two concurrent mutations both read the same document, both render
// their own version and the second os.Rename wins — the first record is gone
// with no trace, because this feature's auto-commit then leaves a clean tree.
// The lock is what makes the "the second sees the first's content" promise true
// rather than aspirational.
//
// A timeout is returned as an error and the mutation does NOT happen. That is
// deliberate: exiting non-zero without writing is honest and repeatable,
// whereas proceeding would reintroduce exactly the clobber this prevents.
func lockRoadmapState(path string) (func(), error) {
	release, err := filelock.Acquire(stateLockPath(path),
		filelock.DefaultTimeout, filelock.DefaultPollInterval,
		"another roadmap mutation is in flight; nothing was written — retry")
	if err != nil {
		return nil, fmt.Errorf("roadmap %w", err)
	}
	return release, nil
}
