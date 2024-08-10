package elasticxmsearch

import "github.com/elastic/go-elasticsearch/v8/typedapi/types"

type Item struct {
	Header types.MultisearchHeader
	Body   types.MultisearchBody
}
