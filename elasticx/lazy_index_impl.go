package elasticx

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type lazyIndex struct {
	idx    *index
	exists *bool
}

// newLazyIndex creates a new lazyIndex implementation.
func newLazyIndex(name string, engine *engine) (*lazyIndex, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if engine == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	idx, err := newIndex(name, engine)
	if err != nil {
		return nil, err
	}
	return &lazyIndex{
		idx: idx,
	}, nil
}

func (li *lazyIndex) checkIndexExists(ctx context.Context) error {
	if li.exists != nil {
		if !*li.exists {
			return newIndexNotFoundCliniaError(li.idx.name)
		}
		return nil
	}
	exists, err := li.idx.es.Indices.Exists(li.indexName().String()).Do(ctx)
	li.exists = &exists
	if err != nil || !exists {
		return newIndexNotFoundCliniaError(li.idx.name)
	}
	return nil
}

// indexName creates the relative path to this index (`clinia-engine~<engine-name>~<name>`)
func (li *lazyIndex) indexName() IndexName {
	return li.idx.indexName()
}

// All of the following methods do not check if the index exists before calling the underlying index method, as the underlying index method will return an error if the index does not exist, which will be handled by the lazyIndex's handleError method to return a NotFoundError.
func (li *lazyIndex) Info() IndexInfo {
	return li.idx.Info()
}

func (li *lazyIndex) Engine() Engine {
	return li.idx.Engine()
}

func (li *lazyIndex) Remove(ctx context.Context) error {
	return li.idx.Remove(ctx)
}

func (li *lazyIndex) UpdateMappings(ctx context.Context, mappings *types.TypeMapping) error {
	return li.idx.UpdateMappings(ctx, mappings)
}

func (li *lazyIndex) ReadDocument(ctx context.Context, id string, result interface{}) (*DocumentMeta, error) {
	return li.idx.ReadDocument(ctx, id, result)
}

// All of the following methods check if the index exists before calling the underlying index method, since the underlying elastic APIs for these methods do not return an error if the index does not exist, and we want to return a NotFoundError if the index does not exist.
func (li *lazyIndex) DocumentExists(ctx context.Context, id string) (bool, error) {
	if err := li.checkIndexExists(ctx); err != nil {
		return false, err
	}

	return li.idx.DocumentExists(ctx, id)
}

func (li *lazyIndex) CreateDocument(ctx context.Context, document interface{}, opts ...DocumentOption) (*DocumentMeta, error) {
	if err := li.checkIndexExists(ctx); err != nil {
		return nil, err
	}

	return li.idx.CreateDocument(ctx, document, opts...)
}

func (li *lazyIndex) UpsertDocument(ctx context.Context, key string, document interface{}, opts ...DocumentOption) (*UpsertResponse[DocumentMeta], error) {
	if err := li.checkIndexExists(ctx); err != nil {
		return nil, err
	}

	return li.idx.UpsertDocument(ctx, key, document, opts...)
}

func (li *lazyIndex) DeleteDocument(ctx context.Context, key string, opts ...DocumentOption) error {
	if err := li.checkIndexExists(ctx); err != nil {
		return err
	}

	return li.idx.DeleteDocument(ctx, key, opts...)
}

func (li *lazyIndex) DeleteDocumentsByQuery(ctx context.Context, query *types.Query, opts ...DocumentOption) (*DeleteQueryResponse, error) {
	if err := li.checkIndexExists(ctx); err != nil {
		return nil, err
	}

	return li.idx.DeleteDocumentsByQuery(ctx, query, opts...)
}

func (li *lazyIndex) UpdateDocumentsByQuery(ctx context.Context, query *types.Query, updateScript *types.Script, opts ...DocumentOption) (*UpdateQueryResponse, error) {
	if err := li.checkIndexExists(ctx); err != nil {
		return nil, err
	}

	return li.idx.UpdateDocumentsByQuery(ctx, query, updateScript, opts...)
}
