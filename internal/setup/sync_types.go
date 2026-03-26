package setup

type SyncAction string

const (
	SyncCreate       SyncAction = "create"
	SyncUpdate       SyncAction = "update"
	SyncManualReview SyncAction = "manual-review"
)

type SyncKind string

const (
	SyncClaudeHooks  SyncKind = "claude-hooks"
	SyncOpenCodeCfg  SyncKind = "opencode-config"
	SyncOpenCodePlug SyncKind = "opencode-plugin"
	SyncAgents       SyncKind = "agents"
	setupDocVersion           = "1"
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
