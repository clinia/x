package elasticx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/clinia/x/errorx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

const (
	enginesIndexName = "clinia-engines"
)

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
		if exists, err := c.es.Indices.Exists(enginesIndexName).Do(ctx); err != nil {
			return err
		} else if exists {
			return nil
		}

		// Create index
		_, err = c.es.Indices.Create(enginesIndexName).
			Request(&create.Request{
				Mappings: &types.TypeMapping{
					Dynamic: &dynamicmapping.Strict,
					Properties: map[string]types.Property{
						"id":   types.NewKeywordProperty(),
						"name": types.NewKeywordProperty(),
					},
				},
			}).
			Do(ctx)
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
		indices, err := c.es.Cat.Indices().Index(fmt.Sprintf("%s*", enginesIndexName)).Do(ctx)
		if err != nil {
			return err
		}

		// Delete all clinia indices
		for _, index := range indices {
			_, err = c.es.Indices.Delete(*index.Index).Do(ctx)
			if err != nil {
				return err
			}
		}

		// Check if clinia-engines index exists
		// If it does not exist, we are done
		if exists, err := c.es.Indices.Exists(enginesIndexName).Do(ctx); err != nil {
			return err
		} else if !exists {
			return nil
		}

		// Remove all engine info from the server
		_, err = c.es.DeleteByQuery(enginesIndexName).
			Query(&types.Query{
				MatchAll: &types.MatchAllQuery{},
			}).
			Do(ctx)
		if err != nil {
			return err
		}

		return nil
	}, bc)
}

// Engine opens a connection to an existing engine.
// If no engine with given name exists, a NotFoundError is returned.
func (c *client) Engine(ctx context.Context, name string) (Engine, error) {
	_, err := c.es.Get(enginesIndexName, name).Do(ctx)
	if err != nil {
		return nil, err
	}

	return &engine{
		name: name,
		es:   c.es,
	}, nil
}

// EngineExists returns true if an engine with given name exists.
func (c *client) EngineExists(ctx context.Context, name string) (bool, error) {
	return c.es.Exists(enginesIndexName, name).Do(ctx)
}

// Engines returns a list of all engines found by the client.
func (c *client) Engines(ctx context.Context) ([]Engine, error) {
	res, err := c.es.Search().
		Index(enginesIndexName).
		Query(&types.Query{
			MatchAll: &types.MatchAllQuery{},
		}).
		From(0).
		Size(1000).
		Sort(
			types.SortOptions{
				SortOptions: map[string]types.FieldSort{
					"name": {
						Order: &sortorder.Desc,
					},
				},
			},
		).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	engines := []Engine{}
	for _, hit := range res.Hits.Hits {
		var engineInfo EngineInfo
		err := json.Unmarshal(hit.Source_, &engineInfo)
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
func (c *client) CreateEngine(ctx context.Context, name string) (Engine, error) {
	e, err := newEngine(name, c.es)
	if err != nil {
		return nil, err
	}

	_, err = c.es.Create(enginesIndexName, name).
		Document(&EngineInfo{
			Name: name,
		}).
		Refresh(refresh.Waitfor).
		Do(ctx)

	if err != nil {
		if isElasticAlreadyExistsError(err) {
			return nil, errorx.AlreadyExistsErrorf("engine '%s' already exists", name)
		}
		return nil, err
	}

	return e, nil
}

func (c *client) ensureConnection(ctx context.Context) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = time.Minute * 5
	bc.Reset()

	return backoff.Retry(func() error {
		ok, err := c.es.Ping().Do(ctx)

		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("could not connect to elastic")
		}

		return nil
	}, bc)
}
