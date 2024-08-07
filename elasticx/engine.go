package elasticx

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/searchtype"
)

// Engine provides access to all indexes in a single engine.
type Engine interface {
	// Name returns the name of the engine.
	Name() string

	// Info fetches the information about the engine.
	Info(ctx context.Context) (*EngineInfo, error)

	// Remove removes the entire engine.
	// If the engine does not exists, a NotFoundError us returned
	Remove(ctx context.Context) error

	// Index functions
	EngineIndexes

	// Search performs a search request to Elastic Search
	Search(ctx context.Context, query *search.Request, indices ...string) (*search.Response, error)

	// MultiSearch performs a multi search request to Elastic Search
	MultiSearch(ctx context.Context, items []MultiSearchItem, queryParams SearchQueryParams) (*msearch.Response, error)

	// Bulk performs a bulk request to Elastic Search
	Bulk(ctx context.Context, ops []BulkOperation, queryParams BulkQueryParams) (*bulk.Response, error)
}

type BulkQueryParams struct {
	// Refresh If `true`, Elasticsearch refreshes the affected shards to make this operation
	// visible to search, if `wait_for` then wait for a refresh to make this
	// operation visible to search, if `false` do nothing with refreshes.
	// Valid values: `true`, `false`, `wait_for`.
	// API name: refresh
	Refresh *refresh.Refresh

	// Pipeline ID of the pipeline to use to preprocess incoming documents.
	// If the index has a default ingest pipeline specified, then setting the value
	// to `_none` disables the default ingest pipeline for this request.
	// If a final pipeline is configured it will always run, regardless of the value
	// of this parameter.
	// API name: pipeline
	Pipeline *string

	// Routing Custom value used to route operations to a specific shard.
	// API name: routing
	Routing *string

	// Source_ `true` or `false` to return the `_source` field or not, or a list of fields
	// to return.
	// API name: _source
	Source_ *string

	// SourceExcludes_ A comma-separated list of source fields to exclude from the response.
	// API name: _source_excludes
	SourceExcludes_ *[]string

	// SourceIncludes_ A comma-separated list of source fields to include in the response.
	// API name: _source_includes
	SourceIncludes_ *[]string

	// Timeout Explicit operation timeout.
	// API name: timeout
	Timeout *string

	// WaitForActiveShards The number of shard copies that must be active before proceeding with the
	// operation.
	// Set to all or any positive integer up to the total number of shards in the
	// index (`number_of_replicas+1`).
	// API name: wait_for_active_shards
	WaitForActiveShards *string

	// RequireAlias If `true`, the request’s actions must target an index alias.
	// API name: require_alias
	RequireAlias *bool

	// ErrorTrace When set to `true` Elasticsearch will include the full stack trace of errors
	// when they occur.
	// API name: error_trace
	ErrorTrace *bool

	// FilterPath Comma-separated list of filters in dot notation which reduce the response
	// returned by Elasticsearch.
	// API name: filter_path
	FilterPath *[]string

	// Human When set to `true` will return statistics in a format suitable for humans.
	// For example `"exists_time": "1h"` for humans and
	// `"eixsts_time_in_millis": 3600000` for computers. When disabled the human
	// readable values will be omitted. This makes sense for responses being
	// consumed
	// only by machines.
	// API name: human
	Human *bool

	// Pretty If set to `true` the returned JSON will be "pretty-formatted". Only use
	// this option for debugging only.
	// API name: pretty
	Pretty *bool
}

type SearchQueryParams struct {
	AllowNoIndices             *bool
	CcsMinimizeRoundtrips      *bool
	ExpandWildcards            *types.ExpandWildcards
	IgnoreThrottled            *bool
	IgnoreUnavailable          *bool
	MaxConcurrentSearches      *string
	MaxConcurrentShardRequests *string
	PreFilterShardSize         *string
	RestTotalHitsAsInt         *bool
	Routing                    *string
	SearchType                 *searchtype.SearchType
	TypedKeys                  *bool
}

type MultiSearchItem struct {
	Header types.MultisearchHeader
	Body   types.MultisearchBody
}

type BulkOperation struct {
	IndexName  string
	Action     BulkAction
	DocumentID string
	Doc        interface{}
}

type BulkAction string

const (
	BulkActionIndex  BulkAction = "index"
	BulkActionCreate BulkAction = "create"
	BulkActionUpdate BulkAction = "update"
	BulkActionDelete BulkAction = "delete"
)

type EngineInfo struct {
	// The name of the engine.
	Name string `json:"name,omitempty"`
}

// EngineIndexes provides access to all indexes in a single engine.
type EngineIndexes interface {
	// Index opens a connection to an exisiting index within the engine.
	// If no index with given name exists, a NotFoundError is returned.
	Index(ctx context.Context, name string) (Index, error)

	// IndexExists returns true if an index with given name exists within the engine.
	IndexExists(ctx context.Context, name string) (bool, error)

	// Indexes returns a list of all indexes in the engine.
	Indexes(ctx context.Context) ([]IndexInfo, error)

	// CreateIndex creates a new index,
	// with given name, and opens a connection to it.
	CreateIndex(ctx context.Context, name string, options *CreateIndexOptions) (Index, error)
}

// CreateIndexOptions contains options that customize the creation of an index.
type CreateIndexOptions struct {
	Aliases  map[string]types.Alias
	Settings *types.IndexSettings
	Mappings *types.TypeMapping
}
