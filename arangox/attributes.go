package arangox

import "go.opentelemetry.io/otel/attribute"

// Attribute key constants for spans and log fields produced by this package.
// Consumers can use these to filter or enrich telemetry without hard-coding strings.
const (
	AttrPackage = "package"

	AttrDBSystem         = "db.system"
	AttrDBCollectionName = "db.collection.name"

	AttrMigrationDirection             = "migration.direction"
	AttrMigrationDryRun                = "migration.dry_run"
	AttrMigrationTargetVersion         = "migration.target_version"
	AttrMigrationCurrentVersion        = "migration.current_version"
	AttrMigrationLatestVersion         = "migration.latest_version"
	AttrMigrationResolvedTargetVersion = "migration.resolved_target_version"
	AttrMigrationVersion               = "migration.version"
	AttrMigrationDescription           = "migration.description"

	attrDBSystemValue    = "arangodb"
	attrMigrationDirUp   = "up"
	attrMigrationDirDown = "down"
)

func attrDBSystem() attribute.KeyValue {
	return attribute.String(AttrDBSystem, attrDBSystemValue)
}

func attrPackage(pkg string) attribute.KeyValue {
	return attribute.String(AttrPackage, pkg)
}

func attrDBCollectionName(name string) attribute.KeyValue {
	return attribute.String(AttrDBCollectionName, name)
}

func attrMigrationDirectionUp() attribute.KeyValue {
	return attribute.String(AttrMigrationDirection, attrMigrationDirUp)
}

func attrMigrationDirectionDown() attribute.KeyValue {
	return attribute.String(AttrMigrationDirection, attrMigrationDirDown)
}

func attrMigrationDryRun(dryRun bool) attribute.KeyValue {
	return attribute.Bool(AttrMigrationDryRun, dryRun)
}

func attrMigrationTargetVersion(v int) attribute.KeyValue {
	return attribute.Int(AttrMigrationTargetVersion, v)
}

// attrMigrationCurrentVersion converts a uint version to an int attribute.
// Migration version numbers are far below math.MaxInt, so this conversion is safe.
func attrMigrationCurrentVersion(v uint) attribute.KeyValue {
	return attribute.Int(AttrMigrationCurrentVersion, int(v)) //nolint:gosec
}

// attrMigrationLatestVersion converts a uint version to an int attribute.
func attrMigrationLatestVersion(v uint) attribute.KeyValue {
	return attribute.Int(AttrMigrationLatestVersion, int(v)) //nolint:gosec
}

// attrMigrationResolvedTargetVersion converts a uint version to an int attribute.
func attrMigrationResolvedTargetVersion(v uint) attribute.KeyValue {
	return attribute.Int(AttrMigrationResolvedTargetVersion, int(v)) //nolint:gosec
}

// attrMigrationVersion converts a uint version to an int attribute.
func attrMigrationVersion(v uint) attribute.KeyValue {
	return attribute.Int(AttrMigrationVersion, int(v)) //nolint:gosec
}

func attrMigrationDescription(desc string) attribute.KeyValue {
	return attribute.String(AttrMigrationDescription, desc)
}
