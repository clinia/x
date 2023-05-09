package elasticx

import (
	"context"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

func createClient(t *testing.T) (Client, *elasticsearch.Client) {
	url := "http://localhost:9200"

	urlFromEnv := os.Getenv("ELASTIC")
	if len(urlFromEnv) > 0 {
		url = urlFromEnv
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := NewClient(esConfig)
	if err != nil {
		t.Error(err)
	}

	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		t.Error(err)
	}

	return client, es
}

func TestClient(t *testing.T) {
	client, _ := createClient(t)

	ctx := context.Background()

	err := client.Init(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("should create an engine", func(t *testing.T) {
		ctx := context.Background()

		name := "test-create-engine"
		engine, err := client.CreateEngine(ctx, name, nil)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return an engine", func(t *testing.T) {
		name := "test-get-engine"
		e, err := client.CreateEngine(ctx, name, nil)
		assert.NoError(t, err)
		assert.NotNil(t, e)

		engine, err := client.Engine(ctx, name)
		assert.NoError(t, err)

		assert.NotEmpty(t, engine)
		assert.Equal(t, engine.Name(), name)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return exists", func(t *testing.T) {
		name := "test-engine-exist"

		engine, err := client.CreateEngine(ctx, name, nil)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		exists, err := client.EngineExists(ctx, name)
		assert.NoError(t, err)

		assert.True(t, exists)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return all engines", func(t *testing.T) {
		// Setup
		names := []string{
			"test-engines-1",
			"test-engines-2",
			"test-engines-3",
		}
		for _, name := range names {
			engine, err := client.CreateEngine(ctx, name, nil)
			assert.NotNil(t, engine)
			assert.NoError(t, err)
		}

		// Act
		engines, err := client.Engines(ctx)
		assert.NoError(t, err)

		engineNames := []string{}
		for _, engine := range engines {
			engineNames = append(engineNames, engine.Name())
		}

		assert.Contains(t, engineNames, "test-engines-1")
		assert.Contains(t, engineNames, "test-engines-2")
		assert.Contains(t, engineNames, "test-engines-3")

		for _, engine := range engines {
			err := engine.Remove(ctx)
			assert.NoError(t, err)
		}
	})
}
