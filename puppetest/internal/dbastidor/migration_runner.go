package dbastidor

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
)

func RunMigrations(db *sql.DB, migrationFS fs.FS) error {
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("error reading migrations: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, readErr := fs.ReadFile(migrationFS, entry.Name())
		if readErr != nil {
			return fmt.Errorf("failed to read migration file `%s`: %w", entry.Name(), readErr)
		}

		// Execute simple migration
		if _, execErr := db.Exec(string(content)); execErr != nil {
			return fmt.Errorf("failed to execute migration file `%s`: %w", entry.Name(), execErr)
		}
	}
	return nil
}
