// Package migration allows to perform versioned migrations in your ArangoDB.
package migrate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clinia/x/elasticx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

type indexSpecification struct {
	Name string `json:"name"`
}

type versionRecord struct {
	Version     uint64    `json:"version"`
	Description string    `json:"description,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

const defaultMigrationsIndex = "migrations"

// AllAvailable used in "Up" or "Down" methods to run all available migrations.
const AllAvailable = -1

// Migrate is type for performing migrations in provided database.
// Database versioned using dedicated collection.
// Each migration applying ("up" and "down") adds new document to collection.
// This document consists migration version, migration description and timestamp.
// Current database version determined as version in latest added document (biggest "_key") from collection mentioned above.
type Migrator struct {
	engine          elasticx.Engine
	migrations      []Migration
	migrationsIndex string
}

func NewMigrator(engine elasticx.Engine, migrations ...Migration) *Migrator {
	internalMigrations := make([]Migration, len(migrations))
	copy(internalMigrations, migrations)
	return &Migrator{
		engine:          engine,
		migrations:      internalMigrations,
		migrationsIndex: defaultMigrationsIndex,
	}
}

// SetMigrationsIndex replaces name of index for storing migration information.
// By default it is "migrations".
func (m *Migrator) SetMigrationsIndex(name string) {
	m.migrationsIndex = name
}

func (m *Migrator) indexExist(name string) (isExist bool, err error) {
	indexes, err := m.getIndexes()
	if err != nil {
		return false, err
	}

	for _, c := range indexes {
		if name == c.Name {
			return true, nil
		}
	}
	return false, nil
}

func (m *Migrator) createIndexIfNotExist(name string) error {
	exist, err := m.indexExist(name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	_, err = m.engine.CreateIndex(context.Background(), name, &elasticx.CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"version":     types.NewLongNumberProperty(),
				"description": types.NewKeywordProperty(),
				"timestamp":   types.NewDateProperty(),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) getIndexes() (indices []indexSpecification, err error) {
	indexes, err := m.engine.Indexes(context.Background())
	if err != nil {
		return nil, err
	}

	for _, index := range indexes {
		indices = append(indices, indexSpecification{
			Name: index.Name(),
		})
	}

	return
}

// Version returns current engine version and comment.
func (m *Migrator) Version(ctx context.Context) (uint64, string, error) {
	if err := m.createIndexIfNotExist(m.migrationsIndex); err != nil {
		return 0, "", err
	}

	query, err := json.Marshal(&search.Request{
		Query: &types.Query{
			MatchAll: &types.MatchAllQuery{},
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
	})
	if err != nil {
		return 0, "", err
	}

	searchResponse, err := m.engine.Query(ctx, string(query), m.migrationsIndex)

	if err != nil {
		return 0, "", err
	}

	if searchResponse.Hits.Total.Value == 0 {
		return 0, "", err
	}

	var rec versionRecord
	if err := json.Unmarshal(searchResponse.Hits.Hits[0].Source, &rec); err != nil {
		return 0, "", err
	}

	return rec.Version, rec.Description, nil
}

// SetVersion forcibly changes database version to provided.
func (m *Migrator) SetVersion(version uint64, description string) error {
	rec := versionRecord{
		Version:     version,
		Timestamp:   time.Now().UTC(),
		Description: description,
	}

	index, err := m.engine.Index(context.Background(), m.migrationsIndex)
	if err != nil {
		return err
	}

	_, err = index.CreateDocument(context.Background(), rec, nil)
	if err != nil {
		return err
	}

	return nil
}

// Up performs "up" migrations to latest available version.
// If n<=0 all "up" migrations with newer versions will be performed.
// If n>0 only n migrations with newer version will be performed.
func (m *Migrator) Up(ctx context.Context, n int) error {
	currentVersion, _, err := m.Version(ctx)
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := 0, 0; i < len(m.migrations) && p < n; i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Up == nil {
			continue
		}
		p++
		if err := migration.Up(m.engine); err != nil {
			return err
		}
		if err := m.SetVersion(migration.Version, migration.Description); err != nil {
			return err
		}
	}
	return nil
}

// Down performs "down" migration to oldest available version.
// If n<=0 all "down" migrations with older version will be performed.
// If n>0 only n migrations with older version will be performed.
func (m *Migrator) Down(ctx context.Context, n int) error {
	currentVersion, _, err := m.Version(ctx)
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := len(m.migrations)-1, 0; i >= 0 && p < n; i-- {
		migration := m.migrations[i]
		if migration.Version > currentVersion || migration.Down == nil {
			continue
		}
		p++
		if err := migration.Down(m.engine); err != nil {
			return err
		}

		var prevMigration Migration
		if i == 0 {
			prevMigration = Migration{Version: 0}
		} else {
			prevMigration = m.migrations[i-1]
		}
		if err := m.SetVersion(prevMigration.Version, prevMigration.Description); err != nil {
			return err
		}
	}
	return nil
}
