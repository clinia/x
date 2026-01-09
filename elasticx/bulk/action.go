package elasticxbulk

import "github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operationtype"

type Action string

const (
	ActionIndex  Action = "index"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

func (a Action) OperationType() operationtype.OperationType {
	switch a {
	case ActionIndex:
		return operationtype.Index
	case ActionCreate:
		return operationtype.Create
	case ActionUpdate:
		return operationtype.Update
	case ActionDelete:
		return operationtype.Delete
	default:
		return operationtype.OperationType{Name: string(a)}
	}
}
