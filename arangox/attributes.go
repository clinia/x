package arangox

import "go.opentelemetry.io/otel/attribute"

// Attribute key constants — private to this package; use the typed helper
// functions below to construct KeyValues and avoid hard-coded strings.
const (
	packageKey = "package"

	dbSystemKey         = "db.system"
	dbCollectionNameKey = "db.collection.name"

	migrationDirectionKey             = "migration.direction"
	migrationDryRunKey                = "migration.dry_run"
	migrationTargetVersionKey         = "migration.target_version"
	migrationCurrentVersionKey        = "migration.current_version"
	migrationLatestVersionKey         = "migration.latest_version"
	migrationResolvedTargetVersionKey = "migration.resolved_target_version"
	migrationVersionKey               = "migration.version"
	migrationDescriptionKey           = "migration.description"

	dbSystemValue    = "arangodb"
	migrationDirUp   = "up"
	migrationDirDown = "down"
)

// DBSystemAttr returns an attribute identifying the database system.
func DBSystemAttr() attribute.KeyValue {
	return attribute.String(dbSystemKey, dbSystemValue)
}

// PackageAttr returns an attribute for the migration package name.
func PackageAttr(pkg string) attribute.KeyValue {
	return attribute.String(packageKey, pkg)
}

// DBCollectionNameAttr returns an attribute for the database collection name.
func DBCollectionNameAttr(name string) attribute.KeyValue {
	return attribute.String(dbCollectionNameKey, name)
}

// MigrationDirectionUpAttr returns an attribute marking an upward migration direction.
func MigrationDirectionUpAttr() attribute.KeyValue {
	return attribute.String(migrationDirectionKey, migrationDirUp)
}

// MigrationDirectionDownAttr returns an attribute marking a downward migration direction.
func MigrationDirectionDownAttr() attribute.KeyValue {
	return attribute.String(migrationDirectionKey, migrationDirDown)
}

// MigrationDryRunAttr returns an attribute indicating whether the migration is a dry-run.
func MigrationDryRunAttr(dryRun bool) attribute.KeyValue {
	return attribute.Bool(migrationDryRunKey, dryRun)
}

// MigrationTargetVersionAttr returns an attribute for the requested target migration version.
func MigrationTargetVersionAttr(v int) attribute.KeyValue {
	return attribute.Int(migrationTargetVersionKey, v)
}

// MigrationCurrentVersionAttr returns an attribute for the current applied migration version.
// Migration version numbers are far below math.MaxInt, so the uint->int conversion is safe.
func MigrationCurrentVersionAttr(v uint) attribute.KeyValue {
	return attribute.Int(migrationCurrentVersionKey, int(v)) //nolint:gosec
}

// MigrationLatestVersionAttr returns an attribute for the latest known migration version.
func MigrationLatestVersionAttr(v uint) attribute.KeyValue {
	return attribute.Int(migrationLatestVersionKey, int(v)) //nolint:gosec
}

// MigrationResolvedTargetVersionAttr returns an attribute for the clamped resolved target version.
func MigrationResolvedTargetVersionAttr(v uint) attribute.KeyValue {
	return attribute.Int(migrationResolvedTargetVersionKey, int(v)) //nolint:gosec
}

// MigrationVersionAttr returns an attribute for an individual migration's version number.
func MigrationVersionAttr(v uint) attribute.KeyValue {
	return attribute.Int(migrationVersionKey, int(v)) //nolint:gosec
}

// MigrationDescriptionAttr returns an attribute for an individual migration's description.
func MigrationDescriptionAttr(desc string) attribute.KeyValue {
	return attribute.String(migrationDescriptionKey, desc)
}
