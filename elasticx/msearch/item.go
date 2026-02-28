package elasticxmsearch

import "github.com/elastic/go-elasticsearch/v9/typedapi/types"

type Item struct {
	Header types.MultisearchHeader
	Body   types.SearchRequestBody
}
