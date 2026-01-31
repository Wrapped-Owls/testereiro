package puppetest

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners"
)

type Engine struct {
	ts *httptest.Server
	db *DBWrapper
}

func (e *Engine) BaseURL() string {
	if e.ts != nil {
		return e.ts.URL
	}
	return "" // TODO: Check a way to have this URL linked on the engine directly
}

func (e *Engine) DB() *sql.DB {
	if e.db == nil {
		return nil
	}
	return e.db.Connection()
}

func (e *Engine) Teardown() error {
	if e.ts != nil {
		e.ts.Close()
	}
	if e.db != nil && !e.db.IsZero() {
		if dbErr := e.db.Teardown(); dbErr != nil {
			return dbErr
		}
	}

	return nil
}

func (e *Engine) Seed(seeds ...any) error {
	if e.db == nil || e.db.IsZero() {
		return fmt.Errorf("database not initialized")
	}
	for _, s := range seeds {
		if err := dbastidor.ExecuteSeedStruct(e.db.Connection(), s); err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
	}
	return nil
}

func (e *Engine) Execute(t testing.TB, runner runners.Runner) error {
	ctx := stgctx.NewRunnerContext(t.Context())
	return runner.Run(t, ctx)
}
