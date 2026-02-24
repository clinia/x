// Package migration allows to perform versioned migrations in your ArangoDB.
package arangox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	arangoDriver "github.com/arangodb/go-driver"
	"github.com/clinia/x/errorx"
	"github.com/clinia/x/loggerx"
	"github.com/clinia/x/mathx"
	"github.com/clinia/x/otelx"
	"github.com/clinia/x/tracex"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type versionRecord struct {
	Version     uint      `json:"version"`
	Description string    `json:"description,omitempty"`
	Package     string    `json:"package"`
	Timestamp   time.Time `json:"timestamp"`
}

const DefaultMigrationsCollection = "migrations"

// AllAvailable used in "Up" or "Down" methods to run all available migrations.
const AllAvailable = -1

const migratorComponentName = "arangox.Migrator"

// Migrate is type for performing migrations in provided database.
// Database versioned using dedicated collection.
// Each migration applying ("up" and "down") adds new document to collection.
// This document consists migration version, migration description and timestamp.
// Current database version determined as version in latest added document (biggest "_key") from collection mentioned above.
type Migrator struct {
	db                   arangoDriver.Database
	pkg                  string
	dryRun               bool
	l                    *loggerx.Logger
	t                    *otelx.Tracer
	migrations           Migrations
	migrationsCollection string
}

type NewMigratorOptions struct {
	Database   arangoDriver.Database
	Package    string
	Migrations Migrations
	DryRun     bool
	Logger     *loggerx.Logger
	Tracer     *otelx.Tracer
}

func NewMigrator(in NewMigratorOptions) *Migrator {
	internalMigrations := make(Migrations, len(in.Migrations))
	copy(internalMigrations, in.Migrations)
	vers := map[uint]bool{}
	for _, m := range in.Migrations {
		if vers[m.Version] {
			panic(fmt.Sprintf("duplicated migration version %v", m.Version))
		}
		vers[m.Version] = true
	}

	l := in.Logger
	if l == nil {
		l = loggerx.NewDefaultLogger().WithFields(tracex.Component("migrator"))
	}

	t := in.Tracer
	if t == nil {
		t = otelx.NewNoopTracer(migratorComponentName)
	}

	return &Migrator{
		db:                   in.Database,
		pkg:                  in.Package,
		dryRun:               in.DryRun,
		l:                    l.WithFields(PackageAttr(in.Package)),
		t:                    t,
		migrations:           internalMigrations,
		migrationsCollection: DefaultMigrationsCollection,
	}
}

func (m *Migrator) instrument(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *loggerx.Logger) {
	opts = append(opts, trace.WithAttributes(DBSystemAttr(), PackageAttr(m.pkg)))
	return tracex.InstrumentNext(ctx,
		func() *loggerx.Logger { return m.l },
		func(ctx context.Context) *otelx.Tracer { return m.t },
		migratorComponentName,
		name,
		opts...)
}

// SetMigrationsCollection replaces name of collection for storing migration information.
// By default it is "migrations".
func (m *Migrator) SetMigrationsCollection(name string) {
	m.migrationsCollection = name
}

func (m *Migrator) createCollectionIfNotExist(ctx context.Context, name string) (outErr error) {
	ctx, span, l := m.instrument(ctx, "createCollectionIfNotExist", trace.WithAttributes(
		DBSystemAttr(),
		DBCollectionNameAttr(name),
	))
	defer func() {
		statusErrorHandler(span, outErr)
	}()

	exist, err := m.db.CollectionExists(ctx, name)
	if err != nil {
		return err
	}
	if exist {
		l.Debug(ctx, "migrations collection already exists")
		return nil
	} else if m.dryRun {
		err := errorx.FailedPreconditionErrorf("collection %s does not exist", name)
		l.WithError(err).Error(ctx, "when dry-run mode is enabled, we can't create the missing collection")
		return err
	}

	l.Info(ctx, "creating migrations collection")
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

	l.Info(ctx, "migrations collection created")
	return nil
}

// Version returns current database version and comment.
func (m *Migrator) Version(ctx context.Context) (current uint, latest uint, desc string, outErr error) {
	ctx, span, l := m.instrument(ctx, "Version", trace.WithAttributes(
		DBSystemAttr(),
	))
	defer func() {
		if outErr == nil {
			span.SetAttributes(
				MigrationCurrentVersionAttr(current),
				MigrationLatestVersionAttr(latest),
				MigrationDescriptionAttr(desc),
			)
		}
		statusErrorHandler(span, outErr)
	}()

	m.migrations.Sort()
	if len(m.migrations) > 0 {
		latest = m.migrations[len(m.migrations)-1].Version
	}
	if err := m.createCollectionIfNotExist(ctx, m.migrationsCollection); err != nil {
		l.WithError(err).Error(ctx, "failed to ensure migrations collection exists")
		return 0, latest, "", err
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
		l.WithError(err).Error(ctx, "failed to query current migration version")
		return 0, latest, "", err
	}
	defer cursor.Close() //nolint:errcheck,gosec

	var rec versionRecord
	_, err = cursor.ReadDocument(ctx, &rec)
	if err != nil {
		_, ok := err.(arangoDriver.NoMoreDocumentsError)
		if ok {
			l.Error(ctx, "no migration version found, database is at version 0")
			return 0, latest, "", nil
		}
		l.WithError(err).Error(ctx, "failed to read current migration version from cursor")
		return 0, latest, "", err
	}

	l.Debug(ctx, "current migration version retrieved",
		MigrationCurrentVersionAttr(rec.Version),
		MigrationLatestVersionAttr(latest),
	)
	return rec.Version, latest, rec.Description, nil
}

// Up performs "up" migrations up to the specified targetVersion.
// If targetVersion<=0 all "up" migrations will be executed (if not executed yet)
// If targetVersion>0 only migrations where version<=targetVersion will be performed (if not executed yet)
func (m *Migrator) Up(ctx context.Context, targetVersion int) (outErr error) {
	ctx, span, l := m.instrument(ctx, "Up", trace.WithAttributes(
		DBSystemAttr(),
		MigrationDirectionUpAttr(),
		MigrationDryRunAttr(m.dryRun),
		MigrationTargetVersionAttr(targetVersion),
	))
	defer func() {
		statusErrorHandler(span, outErr)
	}()

	m.migrations.Sort()
	currentVersion, latest, _, err := m.Version(ctx)
	if err != nil {
		l.WithError(err).Error(ctx, "failed to get current migration version")
		return err
	}

	latestInt, err := safeUintToInt(latest)
	if err != nil {
		l.WithError(err).Error(ctx, "latest migration version is too large to fit in an int")
		return err
	}

	var target uint
	if targetVersion <= 0 {
		target = latest
	} else {
		target = uint(mathx.Clamp(targetVersion, 0, latestInt)) //nolint:errcheck,gosec
	}

	l.Info(ctx, "running up migrations",
		MigrationCurrentVersionAttr(currentVersion),
		MigrationResolvedTargetVersionAttr(target),
	)

	col, err := m.db.Collection(ctx, m.migrationsCollection)
	if err != nil {
		return err
	}

	for i := 0; i < len(m.migrations); i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Up == nil {
			l.Info(ctx, fmt.Sprintf("migration version %d (%s) already applied or has no up migration, skipping", migration.Version, migration.Description), MigrationVersionAttr(migration.Version), MigrationDescriptionAttr(migration.Description))
			continue
		}

		if migration.Version > target {
			l.Warn(ctx, fmt.Sprintf("migration version %d (%s) is above target version %d, skipping", migration.Version, migration.Description, target), MigrationVersionAttr(migration.Version), MigrationDescriptionAttr(migration.Description))
			break
		}

		if err := func() (outErr error) {
			migAttrs := []attribute.KeyValue{
				MigrationVersionAttr(migration.Version),
				MigrationDescriptionAttr(migration.Description),
			}

			if m.dryRun {
				l.Warn(ctx, fmt.Sprintf("[dry-run] ⬆️ up migration version %d (%s) would be applied for package %s", migration.Version, migration.Description, m.pkg), migAttrs...)
				return nil
			}
			ctx, span, l := m.instrument(ctx, "Up.apply", trace.WithAttributes(migAttrs...))
			defer func() {
				statusErrorHandler(span, outErr)
			}()
			l.Info(ctx, fmt.Sprintf("⬆️ applying up migration version %d (%s)", migration.Version, migration.Description), migAttrs...)

			if err := migration.Up(ctx, m.db); err != nil {
				l.WithError(err).Error(ctx, fmt.Sprintf("❌ failed to apply up migration version %d (%s)", migration.Version, migration.Description))
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
				l.WithError(err).Error(ctx, fmt.Sprintf("❌ failed to record up migration version %d (%s) in database", migration.Version, migration.Description))
				return err
			}

			l.Info(ctx, fmt.Sprintf("✅ up migration version %d (%s) applied", migration.Version, migration.Description), migAttrs...)
			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

// Down performs "down" migration to bring back migrations to `version`.
// If targetVersion<=0 all "down" migrations will be performed.
// If targetVersion>0, only the down migrations where version>targetVersion will be performed (only if they were applied).
func (m *Migrator) Down(ctx context.Context, targetVersion int) (outErr error) {
	ctx, span, l := m.instrument(ctx, "Down", trace.WithAttributes(
		DBSystemAttr(),
		MigrationDirectionDownAttr(),
		MigrationDryRunAttr(m.dryRun),
		MigrationTargetVersionAttr(targetVersion),
	))
	defer func() {
		statusErrorHandler(span, outErr)
	}()

	m.migrations.Sort()
	curVersion, latest, _, err := m.Version(ctx)
	if err != nil {
		l.WithError(err).Error(ctx, "failed to get current migration version")
		return err
	}

	latestInt, err := safeUintToInt(latest)
	if err != nil {
		l.WithError(err).Error(ctx, "latest migration version is too large to fit in an int")
		return err
	}

	version := curVersion
	target := uint(mathx.Clamp(targetVersion, 0, latestInt)) //nolint:errcheck,gosec

	l.Info(ctx, "running down migrations",
		MigrationCurrentVersionAttr(curVersion),
		MigrationResolvedTargetVersionAttr(target),
	)

	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version > version || migration.Down == nil {
			l.Info(ctx, fmt.Sprintf("migration version %d (%s) not applied or has no down migration, skipping", migration.Version, migration.Description), MigrationVersionAttr(migration.Version), MigrationDescriptionAttr(migration.Description))
			continue
		}

		if migration.Version <= uint(target) {
			// We down-ed enough
			l.Warn(ctx, fmt.Sprintf("migration version %d (%s) is at or below target version %d, skipping", migration.Version, migration.Description, target), MigrationVersionAttr(migration.Version), MigrationDescriptionAttr(migration.Description))
			break
		}

		if err := func() (outErr error) {
			migAttrs := []attribute.KeyValue{
				MigrationVersionAttr(migration.Version),
				MigrationDescriptionAttr(migration.Description),
			}

			if m.dryRun {
				l.Warn(ctx, fmt.Sprintf("[dry-run] ⬇️ migration version %d (%s) would be applied for package %s", migration.Version, migration.Description, m.pkg), migAttrs...)
				return nil
			}

			ctx, span, l := m.instrument(ctx, "Down.apply", trace.WithAttributes(migAttrs...))
			defer func() {
				statusErrorHandler(span, outErr)
			}()
			l.Info(ctx, fmt.Sprintf("⬇️ applying down migration version %d (%s)", migration.Version, migration.Description), migAttrs...)

			if err := migration.Down(ctx, m.db); err != nil {
				l.WithError(err).Error(ctx, fmt.Sprintf("❌ failed to apply down migration version %d (%s)", migration.Version, migration.Description))
				return err
			}

			l.Info(ctx, fmt.Sprintf("✅ down migration version %d (%s) applied", migration.Version, migration.Description), migAttrs...)
			return nil
		}(); err != nil {
			return err
		}

		if i == 0 {
			version = 0
		} else {
			version = m.migrations[i-1].Version
		}
	}

	if m.dryRun {
		if target == curVersion {
			l.Warn(ctx, fmt.Sprintf("[dry-run] database version already at '%d', no changes would be applied ✅", curVersion))
		} else {
			l.Warn(ctx, fmt.Sprintf("[dry-run] database version would pass from '%d' to '%d'", curVersion, target))
		}
		return nil
	}

	_, err = m.db.Query(ctx, `
		FOR m IN @@collection 
			FILTER m.package == @pkg AND m.version > @version
			REMOVE m IN @@collection
			LET removed = OLD
  			RETURN removed
	`, map[string]any{
		"@collection": m.migrationsCollection,
		"pkg":         m.pkg,
		"version":     version,
	})
	if err != nil {
		l.WithError(err).Error(ctx, "❌ failed to remove down-ed migration records from database")
		return err
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

func statusErrorHandler(span trace.Span, outErr error) {
	if outErr != nil {
		span.RecordError(outErr)
		span.SetStatus(codes.Error, outErr.Error())
	}
	span.End()
}
