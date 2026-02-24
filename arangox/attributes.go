package arangox

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Attribute key constants — private to this package; use the typed helper
// functions below to construct KeyValues and avoid hard-coded strings.
const (
	packageKey = "package"

	dbSystemKey         = semconv.DBSystemKey
	dbCollectionNameKey = semconv.DBCollectionNameKey

	migrationDirectionKey             attribute.Key = "migration.direction"
	migrationDryRunKey                attribute.Key = "migration.dry_run"
	migrationTargetVersionKey         attribute.Key = "migration.target_version"
	migrationCurrentVersionKey        attribute.Key = "migration.current_version"
	migrationLatestVersionKey         attribute.Key = "migration.latest_version"
	migrationResolvedTargetVersionKey attribute.Key = "migration.resolved_target_version"
	migrationVersionKey               attribute.Key = "migration.version"
	migrationDescriptionKey           attribute.Key = "migration.description"

	dbSystemValue         = "arangodb"
	migrationDirUpValue   = "up"
	migrationDirDownValue = "down"
)

// DBSystemAttr returns an attribute identifying the database system.
func DBSystemAttr() attribute.KeyValue {
	return dbSystemKey.String(dbSystemValue)
}

// PackageAttr returns an attribute for the migration package name.
func PackageAttr(pkg string) attribute.KeyValue {
	return attribute.String(packageKey, pkg)
}

// DBCollectionNameAttr returns an attribute for the database collection name.
func DBCollectionNameAttr(name string) attribute.KeyValue {
	return dbCollectionNameKey.String(name)
}

// MigrationDirectionUpAttr returns an attribute marking an upward migration direction.
func MigrationDirectionUpAttr() attribute.KeyValue {
	return migrationDirectionKey.String(migrationDirUpValue)
}

// MigrationDirectionDownAttr returns an attribute marking a downward migration direction.
func MigrationDirectionDownAttr() attribute.KeyValue {
	return migrationDirectionKey.String(migrationDirDownValue)
}

// MigrationDryRunAttr returns an attribute indicating whether the migration is a dry-run.
func MigrationDryRunAttr(dryRun bool) attribute.KeyValue {
	return migrationDryRunKey.Bool(dryRun)
}

// MigrationTargetVersionAttr returns an attribute for the requested target migration version.
func MigrationTargetVersionAttr(v int) attribute.KeyValue {
	return migrationTargetVersionKey.Int(v)
}

// MigrationCurrentVersionAttr returns an attribute for the current applied migration version.
// Migration version numbers are far below math.MaxInt, so the uint->int conversion is safe.
func MigrationCurrentVersionAttr(v uint) attribute.KeyValue {
	return migrationCurrentVersionKey.Int(int(v)) //nolint:gosec
}

// MigrationLatestVersionAttr returns an attribute for the latest known migration version.
func MigrationLatestVersionAttr(v uint) attribute.KeyValue {
	return migrationLatestVersionKey.Int(int(v)) //nolint:gosec
}

// MigrationResolvedTargetVersionAttr returns an attribute for the clamped resolved target version.
func MigrationResolvedTargetVersionAttr(v uint) attribute.KeyValue {
	return migrationResolvedTargetVersionKey.Int(int(v)) //nolint:gosec
}

// MigrationVersionAttr returns an attribute for an individual migration's version number.
func MigrationVersionAttr(v uint) attribute.KeyValue {
	return migrationVersionKey.Int(int(v)) //nolint:gosec
}

// MigrationDescriptionAttr returns an attribute for an individual migration's description.
func MigrationDescriptionAttr(desc string) attribute.KeyValue {
	return migrationDescriptionKey.String(desc)
}
