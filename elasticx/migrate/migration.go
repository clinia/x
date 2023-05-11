package migrate

import (
	"context"
	"sort"

	"github.com/clinia/x/elasticx"
)

// MigrationFunc is used to define actions to be performed for a migration.
type MigrationFunc func(ctx context.Context, engine elasticx.Engine) error

// Migration represents single engine migration.
// Migration contains:
//
// - version: migration version, must be unique in migration list
//
// - description: text description of migration
//
// - up: callback which will be called in "up" migration process
//
// - down: callback which will be called in "down" migration process for reverting changes
type Migration struct {
	Version     uint64
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

func migrationSort(migrations []Migration) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
}

func HasVersion(migrations []Migration, version uint64) bool {
	for _, m := range migrations {
		if m.Version == version {
			return true
		}
	}
	return false
}
