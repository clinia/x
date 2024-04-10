// Package migration allows to perform versioned migrations in your ArangoDB.
package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/clinia/x/elasticx"
	"github.com/clinia/x/errorx"
	"github.com/clinia/x/mathx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
)

type (
	indexSpecification struct {
		Name string `json:"name"`
	}

	versionRecord struct {
		Version     uint64    `json:"version"`
		Description string    `json:"description,omitempty"`
		Package     string    `json:"package"`
		Timestamp   time.Time `json:"timestamp"`
	}
	migrationVersionInfo struct {
		Version     uint64 `json:"version"`
		Description string `json:"description"`
	}
)

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
	pkg             string
	migrations      Migrations
	migrationsIndex string
}

type NewMigratorOptions struct {
	Engine     elasticx.Engine
	Package    string
	Migrations Migrations
}

func NewMigrator(opts NewMigratorOptions) *Migrator {
	internalMigrations := make([]Migration, len(opts.Migrations))
	copy(internalMigrations, opts.Migrations)
	vers := map[uint64]bool{}
	for _, m := range opts.Migrations {
		if vers[m.Version] {
			panic(fmt.Sprintf("duplicated migration version %v", m.Version))
		}
		vers[m.Version] = true
	}

	// To be valid index name, package name should be replaced with ":".
	pkg := strings.ReplaceAll(opts.Package, "/", ":")

	return &Migrator{
		engine:          opts.Engine,
		pkg:             pkg,
		migrations:      internalMigrations,
		migrationsIndex: defaultMigrationsIndex,
	}
}

// SetMigrationsIndex replaces name of index for storing migration information.
// By default it is "migrations".
func (m *Migrator) SetMigrationsIndex(name string) {
	m.migrationsIndex = name
}

func (m *Migrator) indexExist(ctx context.Context, name string) (isExist bool, err error) {
	indexes, err := m.getIndexes(ctx)
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

func (m *Migrator) createIndexIfNotExist(ctx context.Context, name string) error {
	exist, err := m.indexExist(ctx, name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	_, err = m.engine.CreateIndex(ctx, name, &elasticx.CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"version":     types.NewLongNumberProperty(),
				"description": types.NewKeywordProperty(),
				"package":     types.NewKeywordProperty(),
				"timestamp":   types.NewDateProperty(),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) getIndexes(ctx context.Context) (indices []indexSpecification, err error) {
	indexes, err := m.engine.Indexes(ctx)
	if err != nil {
		return nil, err
	}

	for _, index := range indexes {
		indices = append(indices, indexSpecification{
			Name: index.Name,
		})
	}

	return
}

// Version returns current engine version and comment.
func (m *Migrator) Version(ctx context.Context) (migrationVersionInfo, error) {
	if err := m.createIndexIfNotExist(ctx, m.migrationsIndex); err != nil {
		return migrationVersionInfo{
			Version:     0,
			Description: "",
		}, err
	}

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

	if err != nil {
		return migrationVersionInfo{
			Version:     0,
			Description: "",
		}, err
	}

	if searchResponse.Hits.Total.Value == 0 {
		return migrationVersionInfo{
			Version:     0,
			Description: "",
		}, err
	}

	var rec versionRecord
	if err := json.Unmarshal(searchResponse.Hits.Hits[0].Source_, &rec); err != nil {
		return migrationVersionInfo{
			Version:     0,
			Description: "",
		}, err
	}

	return migrationVersionInfo{
		Version:     rec.Version,
		Description: rec.Description,
	}, nil
}

// setVersion forcibly changes database version to provided.
func (m *Migrator) setVersion(ctx context.Context, version uint64, description string) error {
	rec := versionRecord{
		Version:     version,
		Package:     m.pkg,
		Timestamp:   time.Now().UTC(),
		Description: description,
	}

	index, err := m.engine.Index(ctx, m.migrationsIndex)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s:%d", m.pkg, version)
	exists, err := index.DocumentExists(ctx, id)
	if err != nil {
		return err
	} else if exists {
		pkgStr := ""
		if m.pkg != "" {
			pkgStr = fmt.Sprintf("with package %s and ", m.pkg)
		}
		return errorx.AlreadyExistsErrorf("migration %sversion %v already exists", pkgStr, version)
	}

	_, err = index.ReplaceDocument(ctx, id, rec, elasticx.WithRefresh(refresh.Waitfor))
	if err != nil {
		return err
	}

	return nil
}

// Up performs "up" migrations up to the specified targetVersion.
// If targetVersion<=0 all "up" migrations will be executed (if not executed yet)
// If targetVersion>0 only migrations where version<=targetVersion will be performed (if not executed yet)
func (m *Migrator) Up(ctx context.Context, targetVersion int) error {
	m.migrations.Sort()
	if len(m.migrations) == 0 {
		return nil
	}

	currentVersion, err := m.Version(ctx)
	if err != nil {
		return err
	}

	var target uint64
	if latest := m.migrations[len(m.migrations)-1].Version; targetVersion <= 0 {
		target = latest
	} else {
		target = uint64(mathx.Clamp(targetVersion, 0, int(latest)))
	}

	version := currentVersion.Version

	for i := 0; i < len(m.migrations); i++ {
		migration := m.migrations[i]
		if migration.Version <= version || migration.Up == nil {
			continue
		}

		if migration.Version > target {
			break
		}

		if err := migration.Up(ctx, m.engine); err != nil {
			return err
		}
		if err := m.setVersion(ctx, migration.Version, migration.Description); err != nil {
			return err
		}

		version = migration.Version
	}
	return nil
}

// Down performs "down" migration to oldest available version.
// If n<=0 all "down" migrations with older version will be performed.
// If n>0 only n migrations with older version will be performed.
func (m *Migrator) Down(ctx context.Context, n int) error {
	currentVersion, err := m.Version(ctx)
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	m.migrations.Sort()

	for i, p := len(m.migrations)-1, 0; i >= 0 && p < n; i-- {
		migration := m.migrations[i]
		if migration.Version > currentVersion.Version || migration.Down == nil {
			continue
		}
		p++
		if err := migration.Down(ctx, m.engine); err != nil {
			return err
		}

		var prevMigration Migration
		if i == 0 {
			prevMigration = Migration{Version: 0}
		} else {
			prevMigration = m.migrations[i-1]
		}
		if err := m.setVersion(ctx, prevMigration.Version, prevMigration.Description); err != nil {
			return err
		}
	}
	return nil
}
