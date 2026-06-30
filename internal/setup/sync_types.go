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
}

type SyncPlan struct {
	Items []SyncItem
}

func (p SyncPlan) HasChanges() bool {
	return len(p.Items) > 0
}
