package elasticxbulk

type Operation struct {
	IndexName  string
	Action     Action
	DocumentID string
	Doc        interface{}
}
