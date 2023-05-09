package elasticx

import (
	"context"
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
	Query(ctx context.Context, query string, index ...string) (*SearchResponse, error)

	// Queries performs a multi search request to Elastic Search
	Queries(ctx context.Context, queries ...MultiQuery) (map[string]SearchResponse, error)
}

type EngineInfo struct {
	// The identifier of the engine.
	ID string `json:"id,omitempty"`
	// The name of the engine.
	Name string `json:"name,omitempty"`
}

type MultiQuery struct {
	Name  string
	Query string
	Index []string
}
