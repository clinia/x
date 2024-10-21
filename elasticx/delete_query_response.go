package elasticx

// string or int
type TaskId any

type DeleteQueryResponse struct {
	TaskId      *TaskId
	DeleteCount int64
}
