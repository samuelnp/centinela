package gitutil

// Option is a delivery path the operator may pick for a completed feature.
type Option string

const (
	// OptionPR pushes the branch to origin and opens a pull request.
	OptionPR Option = "pr"
	// OptionMerge delegates to the existing local-merge flow.
	OptionMerge Option = "merge"
)

// DeliveryOptions encodes the delivery matrix. PR needs an origin remote to
// push to; local merge needs a worktree branch to merge. The directive and
// `deliver`'s --via guard both consult this so they can never disagree:
//
//	origin & worktree -> [pr, merge]
//	origin & !worktree -> [pr]
//	!origin & worktree -> [merge]
//	neither -> [] (no delivery target detected)
func DeliveryOptions(hasOrigin, worktreeMode bool) []Option {
	var opts []Option
	if hasOrigin {
		opts = append(opts, OptionPR)
	}
	if worktreeMode {
		opts = append(opts, OptionMerge)
	}
	return opts
}

// Supports reports whether v is one of the offered options.
func Supports(opts []Option, v Option) bool {
	for _, o := range opts {
		if o == v {
			return true
		}
	}
	return false
}
