package elasticx

import "context"

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
	CreateDocument(ctx context.Context, document interface{}, options *DocumentOptions) (*DocumentMeta, error)

	// ReplaceDocument replaces a single document with given key in the collection with the document given in the document argument.
	// The document metadata is returned.
	// If no document exists with given key, a NotFoundError is returned.
	ReplaceDocument(ctx context.Context, key string, document interface{}, options *DocumentOptions) (*DocumentMeta, error)

	// DeleteDocument deletes a single document with given key in the collection.
	// No error is returned when the document is successfully deleted.
	// If no document exists with given key, a NotFoundError is returned.
	DeleteDocument(ctx context.Context, key string, options *DocumentOptions) error
}

type DocumentOptions struct {
	Refresh string
}
