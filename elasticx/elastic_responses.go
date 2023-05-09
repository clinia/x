package elasticx

import "encoding/json"

type SearchResponse struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     json.RawMessage `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

type DeleteByQueryResponse struct {
	Took             int  `json:"took"`
	TimedOut         bool `json:"timed_out"`
	Total            int  `json:"total"`
	Deleted          int  `json:"deleted"`
	Batches          int  `json:"batches"`
	VersionConflicts int  `json:"version_conflicts"`
	Noops            int  `json:"noops"`
	Retries          struct {
		Bulk   int `json:"bulk"`
		Search int `json:"search"`
	}
	ThrottledMillis      int               `json:"throttled_millis"`
	RequestsPerSecond    float32           `json:"requests_per_second"`
	ThrottledUntilMillis int               `json:"throttled_until_millis"`
	Failures             []json.RawMessage `json:"failures"`
}
