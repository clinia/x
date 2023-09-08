package elasticx

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
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

	// Query performs a search request to Elastic Search
	Query(ctx context.Context, query *search.Request, indices ...string) (*search.Response, error)

	// Queries performs a multi search request to Elastic Search
	Queries(ctx context.Context, queries ...MultiQuery) (*msearch.Response, error)

	// Bulk performs a bulk request to Elastic Search
	Bulk(ctx context.Context, ops []BulkOperation) (*bulk.Response, error)
}

type MultiQuery struct {
	IndexName string
	Request   types.MultisearchBody
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
