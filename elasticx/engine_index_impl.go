package elasticx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/wesovilabs/koazee"
)

// Index opens a connection to an exisiting index within the engine.
// If no index with given name exists, a NotFoundError is returned.
func (e *engine) Index(ctx context.Context, name string) (Index, error) {
	escapedName := pathEscape(name)
	indexPath := join(e.relPath(), escapedName)

	res, err := esapi.IndicesGetRequest{
		Index: []string{indexPath},
	}.Do(ctx, e.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	index, err := newIndex(name, e)
	if err != nil {
		return nil, err
	}

	return index, nil
}

// IndexExists returns true if an index with given name exists within the engine.
func (e *engine) IndexExists(ctx context.Context, name string) (bool, error) {
	escapedName := pathEscape(name)
	indexPath := join(e.relPath(), escapedName)

	res, err := esapi.IndicesExistsRequest{
		Index: []string{indexPath},
	}.Do(ctx, e.es)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	} else if res.StatusCode == 404 {
		return false, nil
	}

	return false, withElasticError(res)
}

// Indexes returns a list of all indexes in the engine.
func (e *engine) Indexes(ctx context.Context) ([]Index, error) {
	indexPath := join(e.relPath(), "*")

	res, err := esapi.CatIndicesRequest{
		Index:  []string{indexPath},
		Format: "json",
	}.Do(ctx, e.es)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, withElasticError(res)
	}
	defer res.Body.Close()

	var catIndices []CatIndex
	if err := json.NewDecoder(res.Body).Decode(&catIndices); err != nil {
		return nil, err
	}

	indexes := []Index{}

	for _, catIndex := range catIndices {
		name := koazee.StreamOf(split(catIndex.Index)).Do().Last().String()

		indexes = append(indexes, &index{
			name:   name,
			engine: e,
			es:     e.es,
		})
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

	exists, err := e.IndexExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("duplicate index with name %s", name)
	}

	request := &create.Request{}
	if options != nil {

		aliases := map[string]types.Alias{}

		if len(options.Aliases) > 0 {
			for key, alias := range options.Aliases {
				aliases[join(e.relPath(), string(key))] = alias
			}
		}

		request.Aliases = aliases
		request.Settings = options.Settings
		request.Mappings = options.Mappings
	}

	res, err := esapi.IndicesCreateRequest{
		Index: index.relPath(),
		Body:  esutil.NewJSONReader(request),
	}.Do(ctx, e.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	return index, nil
}
