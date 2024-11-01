package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatestMigrationVersion(t *testing.T) {
	m := NewMigrator(NewMigratorOptions{
		Migrations: []Migration{
			{
				Version:     3,
				Description: "Test third migration",
			},
			{
				Version:     1,
				Description: "Test initial migration",
			},
			{
				Version:     2,
				Description: "Test second migration",
			},
		},
	})

	assert.Equal(t, uint(3), m.LatestMigrationVersion())
}
