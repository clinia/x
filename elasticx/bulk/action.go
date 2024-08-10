package elasticxbulk

type Action string

const (
	ActionIndex  Action = "index"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)
