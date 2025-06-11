package elasticx

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	elasticxbulk "github.com/clinia/x/elasticx/bulk"
	elasticxmsearch "github.com/clinia/x/elasticx/msearch"
	elasticxsearch "github.com/clinia/x/elasticx/search"
	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pointerx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operationtype"
)

var isValidEngineName = regexp.MustCompile("^[a-zA-Z0-9_-]*$").MatchString

// engine implementes the Engine interface
type engine struct {
	name string
	es   *elasticsearch.TypedClient
}

// newEngine creates a new Engine implementation.
func newEngine(name string, es *elasticsearch.TypedClient) (Engine, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	if !isValidEngineName(name) {
		return nil, fmt.Errorf("name contains unauthorized characters")
	}

	if es == nil {
		return nil, fmt.Errorf("elastic client is nil")
	}

	return &engine{
		name: name,
		es:   es,
	}, nil
}

// Name returns the of the engine.
func (e *engine) Name() string {
	return e.name
}

// Info fetches the information about the engine.
func (e *engine) Info(ctx context.Context) (*EngineInfo, error) {
	res, err := e.es.Get(enginesIndexNameSegment, e.name).Do(ctx)
	if err != nil {
		return nil, err
	}

	var engineInfo EngineInfo
	if err := json.Unmarshal(res.Source_, &engineInfo); err != nil {
		return nil, err
	}

	return &engineInfo, nil
}

// Remove removes the entire engine.
// If the engine does not exists, a NotFoundError is returned
func (e *engine) Remove(ctx context.Context) error {
	// Fetch all clinia engine indices
	indexName := NewIndexName(enginesIndexNameSegment, pathEscape(e.name), "*").String()
	indices, err := e.es.Cat.Indices().Index(indexName).Do(ctx)
	if err != nil {
		return err
	}

	// Delete all engine indices
	for _, catIndex := range indices {
		/*
			    Ignoring the error here as catching and returning an error will obstruct the control flow as there is
				no rollback option on failure in Elastic
		*/
		e.es.Indices.Delete(*catIndex.Index).Do(ctx) // nolint:gosec
	}

	// Remove engine info from server
	res, err := e.es.DeleteByQuery(enginesIndexNameSegment).
		Query(&types.Query{
			Term: map[string]types.TermQuery{
				"name": {
					Value: e.name,
				},
			},
		}).
		Do(ctx)
	if err != nil {
		return err
	}

	if res.Deleted != nil && *res.Deleted != 1 {
		return errorx.InternalErrorf("failed to delete the engine info with name '%s' inside the server", e.name)
	}

	return nil
}

func (e *engine) Search(ctx context.Context, request *search.Request, indices []string, opts ...elasticxsearch.Option) (*search.Response, error) {
	if request == nil {
		return nil, errorx.InvalidArgumentErrorf("request is nil")
	}

	if request.From != nil && *request.From < 0 {
		return nil, errorx.InvalidArgumentErrorf("invalid search request: 'from' must be greater than or equal to 0")
	}

	if request.Size != nil && *request.Size < 0 {
		return nil, errorx.InvalidArgumentErrorf("invalid search request: 'size' must be greater than or equal to 0")
	}

	indexPaths := []string{}
	for _, name := range indices {
		indexPaths = append(indexPaths, NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(name)).String())
	}

	s := e.es.Search().Index(strings.Join(indexPaths, ",")).Request(request)

	for _, opt := range opts {
		opt(s)
	}

	return s.Do(ctx)
}

func (e *engine) MultiSearch(ctx context.Context, queries []elasticxmsearch.Item, opts ...elasticxmsearch.Option) (*msearch.Response, error) {
	// No queries, we protect against empty requests
	if len(queries) == 0 {
		return &msearch.Response{
			Took:      0,
			Responses: make([]types.MsearchResponseItem, 0),
		}, nil
	}

	items := []types.MsearchRequestItem{}
	for _, query := range queries {
		// Build index names
		indices := []string{}
		for _, index := range query.Header.Index {
			indices = append(indices, NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(index)).String())
		}
		query.Header.Index = indices

		// Append header
		items = append(items, query.Header)
		// Append body
		items = append(items, query.Body)
	}

	ms := e.es.Msearch().Request(&items)

	for _, opt := range opts {
		opt(ms)
	}

	res, err := ms.Do(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (e *engine) Bulk(ctx context.Context, actions []elasticxbulk.Operation, opts ...elasticxbulk.Option) (*bulk.Response, error) {
	// No operations, we protect against empty requests
	if len(actions) == 0 {
		return &bulk.Response{
			Took:   0,
			Errors: false,
			Items:  make([]map[operationtype.OperationType]types.ResponseItem, 0),
		}, nil
	}

	request := []any{}
	for _, action := range actions {
		indexName := NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(action.IndexName)).String()
		op := types.OperationContainer{}

		switch action.Action {
		case elasticxbulk.ActionCreate:
			op.Create = &types.CreateOperation{
				Index_: pointerx.Ptr(indexName),
			}
			request = append(request, op)
			request = append(request, action.Doc)
		case elasticxbulk.ActionIndex:
			op.Index = &types.IndexOperation{
				Index_: pointerx.Ptr(indexName),
				Id_:    pointerx.Ptr(action.DocumentID),
			}
			request = append(request, op)
			request = append(request, action.Doc)
		case elasticxbulk.ActionUpdate:
			op.Update = &types.UpdateOperation{
				Index_: pointerx.Ptr(indexName),
				Id_:    pointerx.Ptr(action.DocumentID),
			}
			request = append(request, op)
			// Wrap the document in a map with the key "doc"
			updateDoc := map[string]interface{}{
				"doc": action.Doc,
			}
			request = append(request, updateDoc)
		case elasticxbulk.ActionDelete:
			op.Delete = &types.DeleteOperation{
				Index_: pointerx.Ptr(indexName),
				Id_:    pointerx.Ptr(action.DocumentID),
			}
			request = append(request, op)

		default:
			return nil, errorx.FailedPreconditionErrorf("unknown bulk action '%s'", action.Action)
		}

	}

	bulkReq := e.es.Bulk().Request(&request)

	for _, opt := range opts {
		opt(bulkReq)
	}

	res, err := bulkReq.Do(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range res.Items {
		for _, op := range item {
			op.Index_ = IndexName(op.Index_).Name()
		}
	}

	return res, nil
}

// Index opens a connection to an exisiting index within the engine.
// If no index with given name exists, a NotFoundError is returned.
func (e *engine) Index(ctx context.Context, name string) (Index, error) {
	indexName := NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(name))

	// Check if index exists
	_, err := e.es.Indices.Get(indexName.String()).Do(ctx)
	if err != nil {
		if isElasticNotFoundError(err) {
			return nil, errorx.NotFoundErrorf("index with name '%s' does not exist", name)
		}
		return nil, err
	}

	index, err := newIndex(name, e)
	if err != nil {
		return nil, err
	}

	return index, nil
}

// IndexExists returns true if an index with given name exists within the engine.
func (e *engine) IndexExists(ctx context.Context, name string) (bool, error) {
	indexName := NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(name))
	return e.es.Indices.Exists(indexName.String()).Do(ctx)
}

// Indexes returns a list of all indexes in the engine.
func (e *engine) Indexes(ctx context.Context) ([]IndexInfo, error) {
	indexPath := NewIndexName(enginesIndexNameSegment, pathEscape(e.name), "*").String()
	indices, err := e.es.Cat.Indices().Index(indexPath).Do(ctx)
	if err != nil {
		return nil, err
	}

	indexes := []IndexInfo{}
	for _, catIndex := range indices {
		indexes = append(indexes, IndexInfo{IndexName(*catIndex.Index).Name()})
	}

	return indexes, nil
}

// CreateIndex creates a new index,
// with given name, and opens a connection to it.
func (e *engine) CreateIndex(ctx context.Context, name string, options *CreateIndexOptions) (Index, error) {
	index, err := newIndex(name, e)
	if err != nil {
		return nil, err
	}

	request := &create.Request{}
	if options != nil {

		aliases := map[string]types.Alias{}

		if len(options.Aliases) > 0 {
			for key, alias := range options.Aliases {
				aliases[NewIndexName(enginesIndexNameSegment, pathEscape(e.name), pathEscape(key)).String()] = alias
			}
		}

		request.Aliases = aliases
		request.Settings = options.Settings
		request.Mappings = options.Mappings
	}

	_, err = e.es.Indices.Create(index.indexName().String()).Request(request).Do(ctx)
	if err != nil {
		if isElasticAlreadyExistsError(err) {
			return nil, errorx.AlreadyExistsErrorf("duplicate index with name '%s'", name)
		}
		return nil, err
	}

	return index, nil
}
