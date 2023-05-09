package elasticx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/clinia/x/stringsx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

const pathSeparator = "~"

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

func join(elem ...string) string {
	return strings.Join(elem, pathSeparator)
}

func split(name string) []string {
	return strings.Split(name, pathSeparator)
}

func (e *engine) relPath() string {
	escapedName := pathEscape(e.name)
	return join(".clinia-engine", escapedName)
}

// Name returns the of the engine.
func (e *engine) Name() string {
	return e.name
}

// Info fetches the information about the engine.
func (e *engine) Info(ctx context.Context) (*EngineInfo, error) {
	res, err := e.es.Search().
		Index(enginesIndexName).
		Request(&search.Request{
			Query: &types.Query{
				Term: map[string]types.TermQuery{
					"name": {
						Value: e.name,
					},
				},
			},
		}).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	if res.Hits.Total.Value != 1 {
		return nil, errors.New("engine not found")
	}

	var engineInfo EngineInfo
	if err := json.Unmarshal(res.Hits.Hits[0].Source_, &engineInfo); err != nil {
		return nil, err
	}

	return &engineInfo, nil
}

// Remove removes the entire engine.
// If the engine does not exists, a NotFoundError us returned
func (e *engine) Remove(ctx context.Context) error {
	// Fetch all clinia engine indices
	catIndicesResponse, err := esapi.CatIndicesRequest{
		Index:  []string{e.relPath() + "*"},
		Format: "json",
	}.Do(ctx, e.es)

	if err != nil {
		return err
	}

	defer catIndicesResponse.Body.Close()

	if catIndicesResponse.IsError() {
		return withElasticError(catIndicesResponse)
	}

	var catIndices []CatIndex
	if err := json.NewDecoder(catIndicesResponse.Body).Decode(&catIndices); err != nil {
		return err
	}

	// Delete all engine indices
	indexNames := []string{}
	for _, catIndex := range catIndices {
		indexNames = append(indexNames, catIndex.Index)
	}

	if len(indexNames) > 0 {
		indicesDeleteResponse, err := esapi.IndicesDeleteRequest{
			Index: indexNames,
		}.Do(ctx, e.es)

		if err != nil {
			return err
		}

		defer indicesDeleteResponse.Body.Close()

		if indicesDeleteResponse.IsError() {
			return withElasticError(indicesDeleteResponse)
		}
	}

	// Remove engine info from server
	deleteQuery := fmt.Sprintf(`{
		"query": {
			"term": {
				"name": %q 
			}
		}
	}`, e.name)
	wait := true
	refresh := true
	dbqr, err := esapi.DeleteByQueryRequest{
		Index:             []string{enginesIndexName},
		Body:              strings.NewReader(deleteQuery),
		WaitForCompletion: &wait,
		Refresh:           &refresh,
	}.Do(ctx, e.es)

	if err != nil {
		return err
	}

	defer dbqr.Body.Close()

	if dbqr.IsError() {
		return withElasticError(dbqr)
	}

	var deleteByQueryResponse DeleteByQueryResponse
	if err := json.NewDecoder(dbqr.Body).Decode(&deleteByQueryResponse); err != nil {
		return err
	}

	if deleteByQueryResponse.Deleted != 1 {
		return errors.New("failed to delete the engine info inside the server")
	}

	return nil
}

func (e *engine) Query(ctx context.Context, query string, indices ...string) (*SearchResponse, error) {
	indexPaths := []string{}
	for _, i := range indices {
		indexPaths = append(indexPaths, join(e.relPath(), i))
	}

	res, err := esapi.SearchRequest{
		Index: indexPaths,
		Body:  strings.NewReader(query),
	}.Do(ctx, e.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

func (e *engine) Queries(ctx context.Context, queries ...MultiQuery) (map[string]SearchResponse, error) {

	requests := []string{}

	for _, query := range queries {
		// Append header
		requests = append(requests, stringsx.SingleLine(fmt.Sprintf(`{
			"index":%q
		}`, join(e.relPath(), strings.Join(query.Index, ",")))))

		// Append body
		requests = append(requests, stringsx.SingleLine(query.Query))
	}

	body := strings.Join(requests, "\n") + "\n"
	res, err := esapi.MsearchRequest{
		Body: strings.NewReader(body),
	}.Do(ctx, e.es)

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, withElasticError(res)
	}

	type mSearchResponse struct {
		Responses []SearchResponse `json:"responses"`
	}

	var mresponse mSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&mresponse); err != nil {
		return nil, err
	}

	results := map[string]SearchResponse{}

	for i := 0; i < len(mresponse.Responses); i++ {
		query := queries[i]
		results[query.Name] = mresponse.Responses[i]
	}

	return results, nil
}
