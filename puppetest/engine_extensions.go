package puppetest

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

func WithTestServer(handler http.Handler) EngineExtension {
	return func(e *Engine) error {
		e.ts = httptest.NewServer(handler)
		return nil
	}
}

func WithMigrationRunner(migrations fs.FS) EngineExtension {
	return func(e *Engine) error {
		if e.db == nil {
			return fmt.Errorf("database not initialized")
		}
		return dbastidor.RunMigrations(e.db.Connection(), migrations)
	}
}
