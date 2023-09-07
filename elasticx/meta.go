package elasticx

type DocumentMeta struct {
	ID      string `json:"_id"`
	Index   string `json:"_index"`
	Version int64  `json:"_version"`
}
