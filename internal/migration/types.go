package migration

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
)

type Plan struct {
	Items []Item
}

func (p Plan) HasChanges() bool {
	return len(p.Items) > 0
}

type Item struct {
	Path                   string
	Action                 Action
	FromVersion            string
	ToVersion              string
	PreservedKeepBlocks    int
	PreservedCustomSection int
	content                string
}
