package migrate

import (
	"context"
	"testing"

	"github.com/clinia/x/assertx"
	"github.com/clinia/x/elasticx"
	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should add migrations index if it does not exist", func(t *testing.T) {
		engine, err := client.CreateEngine(ctx, "test-migrations-a")
		assert.NoError(t, err)
		engines = append(engines, engine.Name())

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

		exists, err := f.es.Indices.Exists("clinia-engines~test-migrations-a~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		res, err := f.es.Get("clinia-engines~test-migrations-a~migrations", "1").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, res.Found)
		assertx.EqualAsJSONExcept(t, versionRecord{
			Version:     uint64(1),
			Description: "Test initial migration",
		}, res.Source_, []string{"timestamp"})
	})

	t.Cleanup(func() {
		for _, engine := range engines {
			f.cleanEngine(t, engine)
		}
	})
}
