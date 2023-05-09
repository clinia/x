package elasticx

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// EngineIndexes provides access to all indexes in a single engine.
type EngineIndexes interface {
	// Index opens a connection to an exisiting index within the engine.
	// If no index with given name exists, a NotFoundError is returned.
	Index(ctx context.Context, name string) (Index, error)

	// IndexExists returns true if an index with given name exists within the engine.
	IndexExists(ctx context.Context, name string) (bool, error)

	// Indexes returns a list of all indexes in the engine.
	Indexes(ctx context.Context) ([]Index, error)

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
