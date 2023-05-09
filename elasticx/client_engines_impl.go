package elasticx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/segmentio/ksuid"
)

// Engine opens a connection to an existing engine.
// If no engine with given name exists, a NotFoundError is returned.
func (c *client) Engine(ctx context.Context, name string) (Engine, error) {
	query := fmt.Sprintf(`{
		"query": {
			"term": {
				"name": %q 
			}
		}
	}`, name)

	res, err := esapi.SearchRequest{
		Index: []string{enginesIndexName},
		Body:  strings.NewReader(query),
	}.Do(ctx, c.es)

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

	if searchResponse.Hits.Total.Value != 1 {
		return nil, errors.New("engine not found")
	}

	var engineInfo EngineInfo
	if err := json.Unmarshal(searchResponse.Hits.Hits[0].Source, &engineInfo); err != nil {
		return nil, err
	}

	return &engine{
		name: engineInfo.Name,
		es:   c.es,
	}, nil
}

// EngineExists returns true if an engine with given name exists.
func (c *client) EngineExists(ctx context.Context, name string) (bool, error) {
	query := fmt.Sprintf(`{
		"query": {
			"term": {
				"name": %q 
			}
		}
	}`, name)

	res, err := esapi.SearchRequest{
		Index: []string{enginesIndexName},
		Body:  strings.NewReader(query),
	}.Do(ctx, c.es)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return false, withElasticError(res)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return false, err
	}

	if searchResponse.Hits.Total.Value == 1 {
		return true, nil
	}

	return false, nil
}

// Engines returns a list of all engines found by the client.
func (c *client) Engines(ctx context.Context) ([]Engine, error) {
	query := `{
		"from": 0,
		"size": 1000,
		"query": {
			"match_all": {}
		}
	}`

	res, err := esapi.SearchRequest{
		Index: []string{enginesIndexName},
		Body:  strings.NewReader(query),
	}.Do(ctx, c.es)

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

	engines := []Engine{}

	for _, hit := range searchResponse.Hits.Hits {
		var engineInfo EngineInfo
		err := json.Unmarshal(hit.Source, &engineInfo)
		if err != nil {
			return nil, err
		}

		engines = append(engines, &engine{
			name: engineInfo.Name,
			es:   c.es,
		})
	}

	return engines, nil
}

// CreateEngine creates a new engine with given name and opens a connection to it.
// If the engine with given name already exists, a DuplicateError is returned.
func (c *client) CreateEngine(ctx context.Context, name string, options *CreateEngineOptions) (Engine, error) {
	e, err := newEngine(name, c.es)
	if err != nil {
		return nil, err
	}

	exists, err := c.EngineExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("duplicate engine with name %s", name)
	}

	res, err := esapi.IndexRequest{
		Index: enginesIndexName,
		Body: esutil.NewJSONReader(&EngineInfo{
			ID:   ksuid.New().String(),
			Name: name,
		}),
		Refresh: "wait_for",
	}.Do(ctx, c.es)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, withElasticError(res)
	}

	return e, nil
}
