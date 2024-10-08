package elasticx

type UpsertResponse struct {
	Result UpsertResult `json:"result"`
	Meta   DocumentMeta `json:"meta"`
}

type UpsertResult string

const (
	UpsertResultCreated UpsertResult = "created"
	UpsertResultUpdated UpsertResult = "updated"
)

func (ur UpsertResult) String() string {
	return string(ur)
}
