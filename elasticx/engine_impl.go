package elasticx

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/clinia/x/errorx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
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
	res, err := e.es.Get(enginesIndexName, e.name).Do(ctx)

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
	indexName := NewIndexName(enginesIndexName, pathEscape(e.name), "*").String()
	indices, err := e.es.Cat.Indices().Index(indexName).Do(ctx)
	if err != nil {
		return err
	}

	// Delete all engine indices
	for _, catIndex := range indices {
		e.es.Indices.Delete(*catIndex.Index).Do(ctx)
	}

	// Remove engine info from server
	res, err := e.es.DeleteByQuery(enginesIndexName).
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

func (e *engine) Query(ctx context.Context, request *search.Request, indices ...string) (*search.Response, error) {
	indexPaths := []string{}
	for _, name := range indices {
		indexPaths = append(indexPaths, NewIndexName(enginesIndexName, pathEscape(e.name), pathEscape(name)).String())
	}

	return e.es.Search().
		Index(strings.Join(indexPaths, ",")).
		Request(request).
		Do(ctx)
}

func (e *engine) Queries(ctx context.Context, queries ...MultiQuery) (*msearch.Response, error) {
	items := []types.MsearchRequestItem{}
	for _, query := range queries {
		// Append header
		items = append(items, types.MultisearchHeader{
			Index: []string{NewIndexName(enginesIndexName, pathEscape(e.name), pathEscape(query.IndexName)).String()},
		})

		// Append body
		items = append(items, query.Request)
	}

	res, err := e.es.Msearch().Request(items).Do(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Index opens a connection to an exisiting index within the engine.
// If no index with given name exists, a NotFoundError is returned.
func (e *engine) Index(ctx context.Context, name string) (Index, error) {
	indexName := NewIndexName(enginesIndexName, pathEscape(e.name), pathEscape(name))

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
	indexName := NewIndexName(enginesIndexName, pathEscape(e.name), pathEscape(name))
	return e.es.Indices.Exists(indexName.String()).Do(ctx)
}

// Indexes returns a list of all indexes in the engine.
func (e *engine) Indexes(ctx context.Context) ([]IndexInfo, error) {
	indexPath := NewIndexName(enginesIndexName, pathEscape(e.name), "*").String()
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
				aliases[NewIndexName(enginesIndexName, pathEscape(e.name), pathEscape(key)).String()] = alias
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

func (e *engine) indexName() IndexName {
	escapedName := pathEscape(e.name)
	return NewIndexName(enginesIndexName, escapedName)
}
