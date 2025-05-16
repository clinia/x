package elasticx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/clinia/x/errorx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/putmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/result"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
)

const (
	elasticStrictDynamicMappingErrorType = "strict_dynamic_mapping_exception"
	elasticIllegalArgumentErrorType      = "illegal_argument_exception"
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

// UpdateMappings updates the index mappings.
// Properties can only be added to the mappings, existing fields cannot be changed.
func (i *index) UpdateMappings(ctx context.Context, mappings *types.TypeMapping) error {
	request := &putmapping.Request{
		Dynamic:    mappings.Dynamic,
		Properties: mappings.Properties,
	}

	_, err := i.es.Indices.PutMapping(i.indexName().String()).Request(request).Do(ctx)
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
		if isElasticNotFoundError(err) {
			return nil, errorx.NotFoundErrorf("document with key '%s' does not exist", id)
		}
		return nil, err
	}

	if !res.Found {
		return nil, errorx.NotFoundErrorf("document with key '%s' does not exist", id)
	}

	err = json.Unmarshal(res.Source_, &result)
	if err != nil {
		return nil, err
	}

	return &DocumentMeta{
		ID:      res.Id_,
		Index:   IndexName(res.Index_).Name(),
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

func (i *index) UpsertDocument(ctx context.Context, key string, document interface{}, opts ...DocumentOption) (*UpsertResponse[DocumentMeta], error) {
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

	return &UpsertResponse[DocumentMeta]{
		Result: UpsertResult(res.Result.Name),
		Meta: DocumentMeta{
			ID:      res.Id_,
			Index:   IndexName(res.Index_).Name(),
			Version: res.Version_,
		},
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
	return NewIndexName(enginesIndexNameSegment, i.engine.name, escapedName)
}

func (i *index) DeleteDocumentsByQuery(ctx context.Context, query *types.Query, opts ...DocumentOption) (*DeleteQueryResponse, error) {
	if query == nil {
		return nil, errorx.InvalidArgumentErrorf("query cannot be nil")
	}

	options := DefaultDocumentOptions
	for _, opt := range opts {
		opt(options)
	}

	// This block is to simulate the behaviour of the `.Refresh` function on the other elastic search lib calls,
	// for some reason the DeleteByQuery takes a boolean that is converted to a string, as the other functions use
	// the direct string
	refresh, refreshErr := strconv.ParseBool(options.refresh.String())
	if refreshErr != nil {
		refresh = false
	}

	res, err := i.es.DeleteByQuery(i.indexName().String()).
		WaitForCompletion(options.waitForCompletion).
		Refresh(refresh).
		Query(query).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	// As soon as one of the query match fails the whole query execution fails, this is to join all the failures
	// into a single error
	if len(res.Failures) > 0 {
		return nil, errors.Join(lo.Map(
			res.Failures,
			func(item types.BulkIndexByScrollFailure, _ int) error {
				return errors.New(item.Type)
			})...)
	}

	var deleteCount int64 = 0
	if res.Deleted != nil {
		deleteCount = *res.Deleted
	}

	var taskId *TaskId
	if res.Task != nil {
		taskId = (*TaskId)(&res.Task)
	}

	return &DeleteQueryResponse{
		DeleteCount: deleteCount,
		TaskId:      taskId,
	}, nil
}
