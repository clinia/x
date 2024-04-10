// Package migration allows to perform versioned migrations in your ArangoDB.
package arangox

import (
	"context"
	"fmt"
	"time"

	arangoDriver "github.com/arangodb/go-driver"
	"github.com/clinia/x/mathx"
)

type collectionSpecification struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type versionRecord struct {
	Version     uint64    `json:"version"`
	Description string    `json:"description,omitempty"`
	Package     string    `json:"package"`
	Timestamp   time.Time `json:"timestamp"`
}

const defaultMigrationsCollection = "migrations"

// AllAvailable used in "Up" or "Down" methods to run all available migrations.
const AllAvailable = -1

// Migrate is type for performing migrations in provided database.
// Database versioned using dedicated collection.
// Each migration applying ("up" and "down") adds new document to collection.
// This document consists migration version, migration description and timestamp.
// Current database version determined as version in latest added document (biggest "_key") from collection mentioned above.
type Migrator struct {
	db                   arangoDriver.Database
	pkg                  string
	migrations           Migrations
	migrationsCollection string
}

type NewMigratorOptions struct {
	Database   arangoDriver.Database
	Package    string
	Migrations Migrations
}

func NewMigrator(in NewMigratorOptions) *Migrator {
	internalMigrations := make(Migrations, len(in.Migrations))
	copy(internalMigrations, in.Migrations)
	vers := map[uint64]bool{}
	for _, m := range in.Migrations {
		if vers[m.Version] {
			panic(fmt.Sprintf("duplicated migration version %v", m.Version))
		}
		vers[m.Version] = true
	}

	return &Migrator{
		db:                   in.Database,
		pkg:                  in.Package,
		migrations:           internalMigrations,
		migrationsCollection: defaultMigrationsCollection,
	}
}

// SetMigrationsCollection replaces name of collection for storing migration information.
// By default it is "migrations".
func (m *Migrator) SetMigrationsCollection(name string) {
	m.migrationsCollection = name
}

func (m *Migrator) createCollectionIfNotExist(ctx context.Context, name string) error {
	exist, err := m.db.CollectionExists(ctx, name)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	col, err := m.db.CreateCollection(ctx, name, nil)
	if err != nil {
		return err
	}

	_, _, err = col.EnsurePersistentIndex(ctx, []string{"version", "package"}, &arangoDriver.EnsurePersistentIndexOptions{
		Unique: true,
	})

	if err != nil {
		return err
	}

	return nil
}

// Version returns current database version and comment.
func (m *Migrator) Version(ctx context.Context) (current uint64, latest uint64, desc string, outErr error) {
	m.migrations.Sort()
	if len(m.migrations) > 0 {
		latest = m.migrations[len(m.migrations)-1].Version
	}
	if err := m.createCollectionIfNotExist(ctx, m.migrationsCollection); err != nil {
		return 0, latest, "", err
	}

	var cursor arangoDriver.Cursor
	cursor, err := m.db.Query(ctx, `
		FOR m IN @@collection 
			FILTER m.package == @pkg 
			SORT m.version DESC 
			LIMIT 1 
			RETURN m
		`, map[string]interface{}{
		"@collection": m.migrationsCollection,
		"pkg":         m.pkg,
	})
	if err != nil {
		return 0, latest, "", err
	}
	defer cursor.Close()

	var rec versionRecord
	_, err = cursor.ReadDocument(ctx, &rec)
	if err != nil {
		_, ok := err.(arangoDriver.NoMoreDocumentsError)
		if ok {
			return 0, latest, "", nil
		}
		return 0, latest, "", err
	}

	return rec.Version, latest, rec.Description, nil
}

// Up performs "up" migrations up to the specified targetVersion.
// If targetVersion<=0 all "up" migrations will be executed (if not executed yet)
// If targetVersion>0 only migrations where version<=targetVersion will be performed (if not executed yet)
func (m *Migrator) Up(ctx context.Context, targetVersion int) error {
	m.migrations.Sort()
	currentVersion, latest, _, err := m.Version(ctx)
	if err != nil {
		return err
	}

	var target uint64
	if targetVersion <= 0 {
		target = latest
	} else {
		target = uint64(mathx.Clamp(targetVersion, 0, int(latest)))
	}

	col, err := m.db.Collection(ctx, m.migrationsCollection)
	if err != nil {
		return err
	}

	for i := 0; i < len(m.migrations); i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Up == nil {
			continue
		}

		if migration.Version > target {
			break
		}

		if err := migration.Up(ctx, m.db); err != nil {
			return err
		}

		rec := versionRecord{
			Version:     migration.Version,
			Package:     m.pkg,
			Timestamp:   time.Now().UTC(),
			Description: migration.Description,
		}

		_, err = col.CreateDocument(ctx, rec)
		if err != nil {
			return err
		}
	}

	return nil
}

// Down performs "down" migration to bring back migrations to `version`.
// If targetVersion<=0 all "down" migrations will be performed.
// If targetVersion>0, only the down migrations where version>targetVersion will be performed (only if they were applied).
func (m *Migrator) Down(ctx context.Context, targetVersion int) error {
	m.migrations.Sort()
	version, latest, _, err := m.Version(ctx)
	if err != nil {
		return err
	}

	target := uint64(mathx.Clamp(targetVersion, 0, int(latest)))

	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version > version || migration.Down == nil {
			continue
		}

		if migration.Version <= uint64(target) {
			break
		}

		if err := migration.Down(ctx, m.db); err != nil {
			return err
		}

		if i == 0 {
			version = 0
		} else {
			version = m.migrations[i-1].Version
		}
	}

	_, err = m.db.Query(ctx, `
		FOR m IN @@collection 
			FILTER m.package == @pkg AND m.version > @version
			REMOVE m IN @@collection
			LET removed = OLD
  			RETURN removed
	`, map[string]interface{}{
		"@collection": m.migrationsCollection,
		"pkg":         m.pkg,
		"version":     version,
	})
	if err != nil {
		return err
	}

	return nil
}
