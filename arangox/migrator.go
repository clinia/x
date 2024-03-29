// Package migration allows to perform versioned migrations in your ArangoDB.
package arangox

import (
	"context"
	"fmt"
	"time"

	arangoDriver "github.com/arangodb/go-driver"
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
	migrations           []Migration
	migrationsCollection string
}

type NewMigratorOptions struct {
	Database   arangoDriver.Database
	Package    string
	Migrations []Migration
}

func NewMigrator(in NewMigratorOptions) *Migrator {
	internalMigrations := make([]Migration, len(in.Migrations))
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
func (m *Migrator) Version(ctx context.Context) (uint64, string, error) {
	if err := m.createCollectionIfNotExist(ctx, m.migrationsCollection); err != nil {
		return 0, "", err
	}

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
		return 0, "", err
	}
	defer cursor.Close()

	var rec versionRecord
	_, err = cursor.ReadDocument(ctx, &rec)
	if err != nil {
		_, ok := err.(arangoDriver.NoMoreDocumentsError)
		if ok {
			return 0, "", nil
		}
	}

	return rec.Version, rec.Description, nil
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

	col, err := m.db.Collection(ctx, m.migrationsCollection)
	if err != nil {
		return err
	}

	for i, p := 0, 0; i < len(m.migrations) && p < n; i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Up == nil {
			continue
		}
		p++
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

	version := currentVersion
	for i, p := len(m.migrations)-1, 0; i >= 0 && p < n; i-- {
		migration := m.migrations[i]
		if migration.Version > currentVersion || migration.Down == nil {
			continue
		}
		p++
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
