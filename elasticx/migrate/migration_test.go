package migrate

import (
	"context"
	"testing"

	"github.com/clinia/x/elasticx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	ctx := context.Background()

	// It takes ELASTICSEARCH_URL from environment variable if it is set.
	// else it uses default value http://localhost:9200
	config := elasticsearch.Config{}

	client, err := elasticx.NewClient(config)
	assert.NoError(t, err)

	esclient, err := elasticsearch.NewTypedClient(config)
	assert.NoError(t, err)

	// Cleanup/Create engine
	err = client.Init(ctx)
	assert.NoError(t, err)

	err = client.Clean(ctx)
	assert.NoError(t, err)

	t.Run("should add migrations index if it does not exist", func(t *testing.T) {
		engine, err := client.CreateEngine(ctx, "test-migrations-a")
		assert.NoError(t, err)

		migrator := NewMigrator(engine, Migration{
			Version:     uint64(1),
			Description: "Test initial migration",
			Up: func(ctx context.Context, engine elasticx.Engine) error {
				_, err := engine.CreateIndex(ctx, "test-index", nil)
				assert.NoError(t, err)

				return nil
			},
			Down: func(ctx context.Context, engine elasticx.Engine) error {
				index, err := engine.Index(ctx, "test-index")
				if err != nil {
					return err
				}

				return index.Remove(ctx)
			},
		})

		err = migrator.Up(ctx, AllAvailable)
		assert.NoError(t, err)

		exists, err := esclient.Indices.Exists("clinia-engines~test-migrations-a~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		esclient.Get("clinia-engines~test-migrations-a~migrations", "1")
	})

	t.Run("should run migrations for a specific engine only", func(t *testing.T) {})
}
