package persistencex

type BulkOperation[T any] struct {
	// Action is the action to perform on the document
	Action BulkAction `json:"action"`

	// ID is the id of the document to perform the action on
	ID string `json:"id"`

	// Body is the document to perform the action on
	Body T `json:"body"`
}

type BulkAction string

const (
	BulkActionCreate = BulkAction("create")
	BulkActionUpsert = BulkAction("upsert")
	BulkActionDelete = BulkAction("delete")
)

type BulkStats struct {
	// HasError indicates if any of the operation failed
	HasError bool
}
