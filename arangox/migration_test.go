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

	client, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	assert.NoError(t, err)

	// Cleanup/Create database
	exists, err := client.DatabaseExists(ctx, "test")
	assert.NoError(t, err)
	if exists {
		db, err := client.Database(ctx, "test")
		assert.NoError(t, err)

		assert.NoError(t, db.Remove(ctx))
	}

	db, err := client.CreateDatabase(ctx, "test", nil)
	assert.NoError(t, err)

	t.Run("should be able to add migration with same version in multiple packages", func(t *testing.T) {
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
		cur, err := db.Query(ctx, `
		FOR m IN @@migrationsCollection
			RETURN m
		`, map[string]interface{}{
			"@migrationsCollection": fMig.migrationsCollection,
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

}
