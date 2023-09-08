package elasticx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/clinia/x/errorx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/result"
	"github.com/segmentio/ksuid"
)

const (
	elasticStrictDynamicMappingErrorType = "strict_dynamic_mapping_exception"
)

type index struct {
	name   string
	engine *engine
	es     *elasticsearch.TypedClient
}

// newIndex creates a new Index implementation.
func newIndex(name string, engine *engine) (*index, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if engine == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	return &index{
		name:   name,
		engine: engine,
		es:     engine.es,
	}, nil
}

// Name returns the name of the index.
func (i *index) Info() IndexInfo {
	return IndexInfo{
		Name: i.name,
	}
}

// Engine returns the engine containing the index.
func (i *index) Engine() Engine {
	return i.engine
}

// Remove removes the entire index.
// If the view does not exists, a NotFoundError us returned.
func (i *index) Remove(ctx context.Context) error {
	_, err := i.es.Indices.Delete(i.indexName().String()).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (i *index) DocumentExists(ctx context.Context, id string) (bool, error) {
	return i.es.Exists(i.indexName().String(), id).Do(ctx)
}

func (i *index) ReadDocument(ctx context.Context, id string, result interface{}) (*DocumentMeta, error) {
	res, err := i.es.Get(i.indexName().String(), id).Do(ctx)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Source_, &result)
	if err != nil {
		return nil, err
	}

	return &DocumentMeta{
		ID:      res.Id_,
		Index:   res.Index_,
		Version: *res.Version_,
	}, nil
}

func (i *index) CreateDocument(ctx context.Context, document interface{}, opts ...DocumentOption) (*DocumentMeta, error) {
	options := DefaultDocumentOptions
	for _, opt := range opts {
		opt(options)
	}

	res, err := i.es.Create(i.indexName().String(), ksuid.New().String()).
		Document(document).
		Refresh(options.refresh).
		Do(ctx)
	if err != nil {
		if eserr, ok := isElasticError(err); ok && eserr.ErrorCause.Type == elasticStrictDynamicMappingErrorType {
			return nil, errorx.FailedPreconditionErrorf("document contains fields that are not allowed by the index mapping: %s", *eserr.ErrorCause.Reason)
		}
		return nil, err
	}

	return &DocumentMeta{
		ID:      res.Id_,
		Index:   IndexName(res.Index_).Name(),
		Version: res.Version_,
	}, nil
}

func (i *index) ReplaceDocument(ctx context.Context, key string, document interface{}, opts ...DocumentOption) (*DocumentMeta, error) {
	options := DefaultDocumentOptions
	for _, opt := range opts {
		opt(options)
	}

	res, err := i.es.Index(i.indexName().String()).
		Id(key).
		Document(document).
		Refresh(options.refresh).
		Do(ctx)
	if err != nil {
		if eserr, ok := isElasticError(err); ok && eserr.ErrorCause.Type == elasticStrictDynamicMappingErrorType {
			return nil, errorx.FailedPreconditionErrorf("document contains fields that are not allowed by the index mapping: %s", *eserr.ErrorCause.Reason)
		}
		return nil, err
	}

	return &DocumentMeta{
		ID:      res.Id_,
		Index:   IndexName(res.Index_).Name(),
		Version: res.Version_,
	}, nil
}

func (i *index) DeleteDocument(ctx context.Context, key string, opts ...DocumentOption) error {
	options := DefaultDocumentOptions
	for _, opt := range opts {
		opt(options)
	}

	res, err := i.es.Delete(i.indexName().String(), key).
		Refresh(options.refresh).
		Do(ctx)
	if err != nil {
		if isElasticNotFoundError(err) {
			return errorx.NotFoundErrorf("document with key '%s' does not exist", key)
		}
		return err
	}

	if res.Result != result.Deleted {
		return errorx.NotFoundErrorf("document with key '%s' does not exist", key)
	}

	return nil
}

// indexName creates the relative path to this index (`clinia-engine~<engine-name>~<name>`)
func (i *index) indexName() IndexName {
	escapedName := pathEscape(i.name)
	return NewIndexName(enginesIndexName, i.engine.name, escapedName)
}
