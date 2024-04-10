package migrate

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/clinia/x/assertx"
	"github.com/clinia/x/elasticx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		migrator := NewMigrator(NewMigratorOptions{
			Engine:  engine,
			Package: "",
			Migrations: []Migration{
				{
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
				},
			},
		})

		err = migrator.Up(ctx, AllAvailable)
		assert.NoError(t, err)

		exists, err := f.es.Indices.Exists("clinia-engines~test-migrations-a~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		res, err := f.es.Get("clinia-engines~test-migrations-a~migrations", ":1").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, res.Found)
		assertx.EqualAsJSONExcept(t, versionRecord{
			Version:     uint64(1),
			Description: "Test initial migration",
		}, res.Source_, []string{"timestamp"})

		verInfo, err := migrator.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, migrationVersionInfo{
			Version:     uint64(1),
			Description: "Test initial migration",
		}, verInfo)
	})

	t.Run("should be able to add migration with same version in multiple packages only", func(t *testing.T) {
		engine, err := client.CreateEngine(ctx, "test-migrations-b")
		assert.NoError(t, err)
		engines = append(engines, engine.Name())

		migratorA := NewMigrator(NewMigratorOptions{
			Engine:  engine,
			Package: "package-1",
			Migrations: []Migration{
				{
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
				},
			},
		})

		migratorB := NewMigrator(NewMigratorOptions{
			Engine:  engine,
			Package: "package-2",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, engine elasticx.Engine) error {
						_, err := engine.CreateIndex(ctx, "test-index-2", nil)
						assert.NoError(t, err)

						return nil
					},
					Down: func(ctx context.Context, engine elasticx.Engine) error {
						index, err := engine.Index(ctx, "test-index-2")
						if err != nil {
							return err
						}

						return index.Remove(ctx)
					},
				},
			},
		})

		err = migratorA.Up(ctx, AllAvailable)
		assert.NoError(t, err)

		err = migratorB.Up(ctx, AllAvailable)
		assert.NoError(t, err)

		exists, err := f.es.Indices.Exists("clinia-engines~test-migrations-b~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		res, err := f.es.Get("clinia-engines~test-migrations-b~migrations", "package-1:1").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, res.Found)
		assertx.EqualAsJSONExcept(t, versionRecord{
			Version:     uint64(1),
			Package:     "package-1",
			Description: "Test initial migration",
		}, res.Source_, []string{"timestamp"})

		res2, err := f.es.Get("clinia-engines~test-migrations-b~migrations", "package-2:1").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, res2.Found)
		assertx.EqualAsJSONExcept(t, versionRecord{
			Version:     uint64(1),
			Package:     "package-2",
			Description: "Test initial migration",
		}, res2.Source_, []string{"timestamp"})
	})

	t.Run("should panic when passing a migrations with conflicting versions", func(t *testing.T) {
		engine, err := client.CreateEngine(ctx, "test-migrations-c")
		assert.NoError(t, err)
		engines = append(engines, engine.Name())

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			} else {
				assert.Equal(t, "duplicated migration version 1", r)
			}
		}()

		// Should panic
		NewMigrator(NewMigratorOptions{
			Engine:  engine,
			Package: "package-1",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
					Down: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
				},
				{
					Version:     uint64(1),
					Description: "Duplicated migration",
					Up: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
					Down: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
				},
			},
		})
	})

	t.Run("should be able to migrate up & down in batches", func(t *testing.T) {
		engine, err := client.CreateEngine(ctx, "test-migrations-d")
		assert.NoError(t, err)
		engines = append(engines, engine.Name())

		m := NewMigrator(NewMigratorOptions{
			Engine:  engine,
			Package: "package-1",
			Migrations: []Migration{
				{
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
				},
				{
					Version:     uint64(2),
					Description: "Test second migration",
					Up: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
					Down: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
				},
				{
					Version:     uint64(3),
					Description: "Test third migration",
					Up: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
					Down: func(ctx context.Context, engine elasticx.Engine) error {
						return nil
					},
				},
			},
		})

		getVersions := func() []versionRecord {
			searchResponse, err := m.engine.Query(ctx, &search.Request{
				Query: &types.Query{
					Term: map[string]types.TermQuery{
						"package": {
							Value: m.pkg,
						},
					},
				},
				Sort: types.Sort{
					types.SortOptions{
						SortOptions: map[string]types.FieldSort{
							"version": {
								Order: &sortorder.Desc,
							},
						},
					},
				},
			}, m.migrationsIndex)

			require.NoError(t, err)

			out := make([]versionRecord, len(searchResponse.Hits.Hits))
			for i, hit := range searchResponse.Hits.Hits {
				var rec versionRecord

				if err := json.Unmarshal(hit.Source_, &rec); err != nil {
					t.Fatal(err)
				}

				out[i] = rec
			}

			return out
		}

		exists, err := f.es.Indices.Exists("clinia-engines~test-migrations-d~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)

		err = m.Up(ctx, 1)
		assert.NoError(t, err)

		exists, err = f.es.Indices.Exists("clinia-engines~test-migrations-d~migrations").Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		versions := getVersions()
		assert.Len(t, versions, 1)
		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))

		err = m.Up(ctx, 3)
		assert.NoError(t, err)

		versions = getVersions()
		assert.Len(t, versions, 3)
		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
			{
				Version:     uint64(2),
				Package:     "package-1",
				Description: "Test second migration",
			},
			{
				Version:     uint64(3),
				Package:     "package-1",
				Description: "Test third migration",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))

		err = m.Down(ctx, 2)
		assert.NoError(t, err)

		versions = getVersions()
		assert.Len(t, versions, 2)
		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
			{
				Version:     uint64(2),
				Package:     "package-1",
				Description: "Test second migration",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))

		err = m.Down(ctx, 0)
		assert.NoError(t, err)

		versions = getVersions()
		assert.Len(t, versions, 0)
	})

	t.Cleanup(func() {
		for _, engine := range engines {
			f.cleanEngine(t, engine)
		}
	})
}
