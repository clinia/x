package elasticx

type UpsertResponse[T any] struct {
	Result UpsertResult `json:"result"`
	Meta   T            `json:"meta"`
}

type UpsertResult string

const (
	UpsertResultCreated UpsertResult = "created"
	UpsertResultUpdated UpsertResult = "updated"
)

func (ur UpsertResult) String() string {
	return string(ur)
}
