package puppetest

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

// WithTestServer attaches an httptest server with the provided handler to the engine.
func WithTestServer(handler http.Handler) EngineExtension {
	return func(e *Engine) error {
		e.ts = httptest.NewServer(handler)
		return nil
	}
}

// WithTestServerFromEngine attaches an httptest server using a handler built from the engine state.
func WithTestServerFromEngine(handlerFactory func(*Engine) (http.Handler, error)) EngineExtension {
	return func(e *Engine) error {
		mainHandler, err := handlerFactory(e)
		if err != nil {
			return fmt.Errorf("could not create main handler: %w", err)
		}
		e.ts = httptest.NewServer(mainHandler)
		return nil
	}
}

// WithMigrationRunner runs SQL migrations against the engine database during extension setup.
func WithMigrationRunner(migrations fs.FS) EngineExtension {
	return func(e *Engine) error {
		if e.db == nil || e.db.IsZero() {
			return fmt.Errorf("database not initialized")
		}
		return dbastidor.RunMigrations(e.db.Connection(), migrations)
	}
}
