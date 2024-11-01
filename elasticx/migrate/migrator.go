// Package migration allows to perform versioned migrations in your ElasticEngine.
package migrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
		Version     uint      `json:"version"`
		Description string    `json:"description,omitempty"`
		Package     string    `json:"package"`
		Timestamp   time.Time `json:"timestamp"`
	}
	migrationVersionInfo struct {
		Version     uint   `json:"version"`
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
	engine                 elasticx.Engine
	pkg                    string
	migrations             Migrations
	migrationsIndex        string
	latestMigrationVersion uint
}

type NewMigratorOptions struct {
	Engine     elasticx.Engine
	Package    string
	Migrations Migrations
}

func NewMigrator(opts NewMigratorOptions) *Migrator {
	internalMigrations := make(Migrations, len(opts.Migrations))
	copy(internalMigrations, opts.Migrations)
	vers := map[uint]bool{}
	for _, m := range opts.Migrations {
		if vers[m.Version] {
			panic(fmt.Sprintf("duplicated migration version %v", m.Version))
		}
		vers[m.Version] = true
	}

	// To be valid index name, package name should be replaced with ":".
	pkg := strings.ReplaceAll(opts.Package, "/", ":")

	// latest migration version
	var latest uint
	if len(internalMigrations) > 0 {
		internalMigrations.Sort()
		latest = internalMigrations[len(internalMigrations)-1].Version
	}

	return &Migrator{
		engine:                 opts.Engine,
		pkg:                    pkg,
		migrations:             internalMigrations,
		migrationsIndex:        defaultMigrationsIndex,
		latestMigrationVersion: latest,
	}
}

// SetMigrationsIndex replaces name of index for storing migration information.
// By default it is "migrations".
func (m *Migrator) SetMigrationsIndex(name string) {
	m.migrationsIndex = name
}

func (m *Migrator) LatestMigrationVersion() uint {
	return m.latestMigrationVersion
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

	searchResponse, err := m.engine.Search(ctx, &search.Request{
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
	}, []string{m.migrationsIndex})
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
func (m *Migrator) setVersion(ctx context.Context, version uint, description string) error {
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

	_, err = index.UpsertDocument(ctx, id, rec, elasticx.WithRefresh(refresh.Waitfor))
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) removeVersion(ctx context.Context, version uint) error {
	index, err := m.engine.Index(ctx, m.migrationsIndex)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s:%d", m.pkg, version)
	err = index.DeleteDocument(ctx, id, elasticx.WithRefresh(refresh.Waitfor))
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

	var target uint
	if latest := m.migrations[len(m.migrations)-1].Version; targetVersion <= 0 {
		target = latest
	} else {
		latestInt, err := safeUintToInt(latest)
		if err != nil {
			return err
		}

		target = uint(mathx.Clamp(targetVersion, 0, latestInt))
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

// Down performs "down" migration to bring back migrations to `version`.
// If targetVersion<=0 all "down" migrations will be performed.
// If targetVersion>0, only the down migrations where version>targetVersion will be performed (only if they were applied).
func (m *Migrator) Down(ctx context.Context, targetVersion int) error {
	m.migrations.Sort()
	if len(m.migrations) == 0 {
		return nil
	}
	currentVersion, err := m.Version(ctx)
	if err != nil {
		return err
	}

	latestVer := m.migrations[len(m.migrations)-1].Version

	latestVerInt, err := safeUintToInt(latestVer)
	if err != nil {
		return err
	}

	target := uint(mathx.Clamp(targetVersion, 0, latestVerInt))
	version := currentVersion.Version

	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version > version || migration.Down == nil {
			continue
		}

		if migration.Version <= target {
			// We down-ed enough
			break
		}

		if err := migration.Down(ctx, m.engine); err != nil {
			return err
		}

		if err := m.removeVersion(ctx, migration.Version); err != nil {
			return err
		}

		if i == 0 {
			version = 0
		} else {
			version = m.migrations[i-1].Version
		}
	}
	return nil
}

func safeUintToInt(u uint) (int, error) {
	if u > math.MaxInt {
		return 0, errors.New("uint value is too large to fit in an int")
	}
	// Suppress gosec warning for safe conversion
	// #nosec G115
	return int(u), nil
}
