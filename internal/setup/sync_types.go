package setup

type SyncAction string

const (
	SyncCreate       SyncAction = "create"
	SyncUpdate       SyncAction = "update"
	SyncManualReview SyncAction = "manual-review"
)

type SyncKind string

const (
	// SyncKindPrewriteHook is the blocking-write hook surface. Claude wires it
	// via settings.json hooks; OpenCode via its plugin file. Adapters that
	// declare CapBlocksWrites must emit an item of this kind.
	SyncKindPrewriteHook SyncKind = "prewrite-hook"
	SyncOpenCodeCfg      SyncKind = "opencode-config"
	SyncAgents           SyncKind = "agents"
	SyncAiderConfig      SyncKind = "aider-config"
	setupDocVersion               = "1"
)

type SyncItem struct {
	Kind   SyncKind
	Path   string
	Action SyncAction
	Reason string
	// Local carries the managed local provider (opencode items only) so apply
	// rewrites the same block the plan computed. nil for every other item and
	// for the zero-config opencode path.
	Local *LocalProvider
}

type SyncPlan struct {
	Items []SyncItem
}

func (p SyncPlan) HasChanges() bool {
	return len(p.Items) > 0
}
