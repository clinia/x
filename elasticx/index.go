package elasticx

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
)

// Index provides access to the information of an index.
type Index interface {
	// Name returns the name of the index.
	Info() IndexInfo

	// Engine returns the engine containing the index.
	Engine() Engine

	// Remove removes the entire index.
	// If the index does not exists, a NotFoundError is returned.
	Remove(ctx context.Context) error

	// All document functions
	IndexDocuments
}

type IndexDocuments interface {
	// DocumentExists checks if a document with given id exists in the index.
	DocumentExists(ctx context.Context, id string) (bool, error)

	// ReadDocument reads a single document with given id from the index.
	// The document data is stored into result, the document metadata is returned.
	// If no document exists with given id, a NotFoundError is returned.
	ReadDocument(ctx context.Context, id string, result interface{}) (*DocumentMeta, error)

	// CreateDocument creates a single document in the index.
	// The document data is loaded from the given document, the document metadata is returned.
	// If the document data already contains a `_key` field, this will be used as key of the new document,
	// otherwise a unique key is created.
	CreateDocument(ctx context.Context, document interface{}, opts ...DocumentOption) (*DocumentMeta, error)

	// ReplaceDocument replaces a single document with given key in the collection with the document given in the document argument.
	// The document metadata is returned.
	// If no document exists with given key, a NotFoundError is returned.
	ReplaceDocument(ctx context.Context, key string, document interface{}, opts ...DocumentOption) (*DocumentMeta, error)

	// DeleteDocument deletes a single document with given key in the collection.
	// No error is returned when the document is successfully deleted.
	// If no document exists with given key, a NotFoundError is returned.
	DeleteDocument(ctx context.Context, key string, opts ...DocumentOption) error
}

type IndexInfo struct {
	Name string `json:"name"`
}

type documentOptions struct {
	refresh refresh.Refresh
}

type DocumentOption func(*documentOptions)

var DefaultDocumentOptions = &documentOptions{
	refresh: refresh.False,
}

// WithRefresh sets the refresh option of the document operation.
func WithRefresh(refresh refresh.Refresh) DocumentOption {
	return func(opts *documentOptions) {
		opts.refresh = refresh
	}
}
