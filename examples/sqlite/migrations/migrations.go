package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var embedMigrationsFS embed.FS

func MigrationFS() fs.FS {
	return embedMigrationsFS
}
