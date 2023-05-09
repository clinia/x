package elasticx

import "context"

// Index provides access to the information of an index.
type Index interface {
	// Name returns the name of the index.
	Name() string

	// Engine returns the engine containing the index.
	Engine() Engine

	// Remove removes the entire index.
	// If the view does not exists, a NotFoundError us returned.
	Remove(ctx context.Context) error

	// All document functions
	IndexDocuments
}
