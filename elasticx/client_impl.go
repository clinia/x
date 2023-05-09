package elasticx

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
)

const (
	enginesIndexName = ".clinia-engines"
)

var engineIndex = &create.Request{
	Mappings: &types.TypeMapping{
		Dynamic: &dynamicmapping.Strict,
		Properties: map[string]types.Property{
			"id":   types.NewKeywordProperty(),
			"name": types.NewKeywordProperty(),
		},
	},
}

type client struct {
	es *elasticsearch.TypedClient
}

// NewClient creates a new Client based on the given config.
func NewClient(config elasticsearch.Config) (Client, error) {
	es, err := elasticsearch.NewTypedClient(config)
	if err != nil {
		return nil, err
	}

	c := &client{
		es: es,
	}

	return c, nil
}

func (c *client) Init(ctx context.Context) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()

	err := c.ensureConnection(ctx)
	if err != nil {
		return err
	}

	return backoff.Retry(func() error {
		// Check if index exists
		res, err := esapi.IndicesExistsRequest{
			Index: []string{enginesIndexName},
		}.Do(ctx, c.es)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		exists := res.StatusCode == 200
		if exists {
			return nil
		}

		_, err = c.es.Indices.Create(enginesIndexName).Request(engineIndex).Do(ctx)
		if err != nil {
			return err
		}

		return nil
	}, bc)
}

func (c *client) Clean(ctx context.Context) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Second * 1
	bc.Reset()

	err := c.ensureConnection(ctx)
	if err != nil {
		return err
	}

	return backoff.Retry(func() error {
		// Fetch all clinia indices
		res, err := esapi.CatIndicesRequest{
			Index:  []string{".clinia-engine~*"},
			Format: "json",
		}.Do(ctx, c.es)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return withElasticError(res)
		}

		var catIndices []CatIndex
		if err := json.NewDecoder(res.Body).Decode(&catIndices); err != nil {
			return err
		}

		indexNames := []string{}
		for _, catIndex := range catIndices {
			indexNames = append(indexNames, catIndex.Index)
		}

		if len(indexNames) > 0 {
			res, err = esapi.IndicesDeleteRequest{
				Index: indexNames,
			}.Do(ctx, c.es)

			if err != nil {
				return err
			}

			defer res.Body.Close()

			if res.IsError() {
				return withElasticError(res)
			}
		}

		// Check if .clinia-engines index exists
		res, err = esapi.IndicesExistsRequest{
			Index: []string{enginesIndexName},
		}.Do(ctx, c.es)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		exists := res.StatusCode == 200
		if !exists {
			return nil
		}

		// Remove all engine info from the server
		deleteQuery := `{
			"query": {
				"match_all": {}
			}
		}`
		dbqr, err := esapi.DeleteByQueryRequest{
			Index: []string{enginesIndexName},
			Body:  strings.NewReader(deleteQuery),
		}.Do(ctx, c.es)

		if err != nil {
			return err
		}

		defer dbqr.Body.Close()

		if dbqr.IsError() {
			return withElasticError(dbqr)
		}

		return nil
	}, bc)
}

func (c *client) ensureConnection(ctx context.Context) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()

	return backoff.Retry(func() error {
		res, err := esapi.PingRequest{}.Do(ctx, c.es)

		if err != nil {
			return err
		}

		if res.IsError() {
			return withElasticError(res)
		}

		return nil
	}, bc)
}
