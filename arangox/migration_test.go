package arangox

import (
	"context"
	"os"
	"testing"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/clinia/x/assertx"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	ctx, db := newFixture(t, "test_migration")

	t.Run("should be able to add migration with same version in multiple packages only", func(t *testing.T) {
		fMig := NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-1",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

		sMig := NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-2",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration for second package",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

		err := fMig.Up(ctx, 0)
		assert.NoError(t, err)

		err = sMig.Up(ctx, 0)
		assert.NoError(t, err)

		// Check if migrations are added to the db
		versions := fetchMigrations(t, ctx, db, fMig)

		assert.Len(t, versions, 2)
		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
			{
				Version:     uint64(1),
				Package:     "package-2",
				Description: "Test initial migration for second package",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))
	})

	t.Run("should panic when passing a migrations with conflicting versions", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			} else {
				assert.Equal(t, "duplicated migration version 1", r)
			}
		}()

		// Should panic
		NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-2",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration for second package",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
				{
					Version:     uint64(1),
					Description: "I should fail since I have the same version within the same package",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

	})

	t.Run("should not be able to create a migration with the same version directly in the database", func(t *testing.T) {
		fMig := NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-3",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

		err := fMig.Up(ctx, 0)
		assert.NoError(t, err)

		col, err := db.Collection(ctx, fMig.migrationsCollection)
		assert.NoError(t, err)

		_, err = col.CreateDocument(ctx, versionRecord{
			Version:     uint64(1),
			Package:     "package-3",
			Description: "Should fail since I have the same version within the same package",
		})

		assert.Error(t, err)
		assert.True(t, driver.IsConflict(err))
	})

}

func TestDownMigrations(t *testing.T) {
	ctx, db := newFixture(t, "test_down-migrations")

	t.Run("should be able to migrate up and down", func(t *testing.T) {
		m := NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-1",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

		err := m.Up(ctx, 0)
		assert.NoError(t, err)

		versions := fetchMigrations(t, ctx, db, m)

		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))

		err = m.Down(ctx, 0)
		assert.NoError(t, err)

		versions = fetchMigrations(t, ctx, db, m)
		assert.Len(t, versions, 0)
	})

	t.Run("should be able to migrate up and down multiple migrations", func(t *testing.T) {
		m := NewMigrator(NewMigratorOptions{
			Database: db,
			Package:  "package-1",
			Migrations: []Migration{
				{
					Version:     uint64(1),
					Description: "Test initial migration",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
				{
					Version:     uint64(2),
					Description: "Test second migration",
					Up: func(ctx context.Context, db driver.Database) error {
						return nil
					},
					Down: func(ctx context.Context, db driver.Database) error {
						return nil
					},
				},
			},
		})

		err := m.Up(ctx, 0)
		assert.NoError(t, err)

		versions := fetchMigrations(t, ctx, db, m)

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

		err = m.Down(ctx, 1)
		assert.NoError(t, err)

		versions = fetchMigrations(t, ctx, db, m)
		assertx.ElementsMatch(t, versions, []versionRecord{
			{
				Version:     uint64(1),
				Package:     "package-1",
				Description: "Test initial migration",
			},
		}, cmpopts.IgnoreFields(versionRecord{}, "Timestamp"))
	})
}

func newFixture(t *testing.T, dbName string) (context.Context, driver.Database) {
	ctx := context.Background()

	dsnHost := "http://localhost:8529"
	dsnFromEnv := os.Getenv("ARANGO_URL")
	if len(dsnFromEnv) > 0 {
		dsnHost = dsnFromEnv
	}

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{dsnHost},
	})
	assert.NoError(t, err)

	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	assert.NoError(t, err)

	// Drop the database if it exists
	exists, err := c.DatabaseExists(ctx, dbName)
	assert.NoError(t, err)
	if exists {
		db, err := c.Database(ctx, dbName)
		assert.NoError(t, err)

		err = db.Remove(ctx)
		assert.NoError(t, err)
	}

	db, err := c.CreateDatabase(ctx, dbName, &driver.CreateDatabaseOptions{})
	assert.NoError(t, err)

	return ctx, db
}

func fetchMigrations(t *testing.T, ctx context.Context, db driver.Database, m *Migrator) []versionRecord {
	// Check if migrations are added to the db
	cur, err := db.Query(ctx, `
	FOR m IN @@migrationsCollection
		RETURN m
	`, map[string]interface{}{
		"@migrationsCollection": m.migrationsCollection,
	})
	assert.NoError(t, err)
	defer cur.Close()

	var versions []versionRecord
	for cur.HasMore() {
		var version versionRecord
		_, err = cur.ReadDocument(ctx, &version)
		assert.NoError(t, err)
		versions = append(versions, version)
	}

	return versions
}
