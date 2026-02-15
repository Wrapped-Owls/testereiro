package puppetest

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/runners"
)

// Context is the internal context object used on the test engine to take some objects from a given state
type Context = stgctx.RunnerContext

type Engine struct {
	ts    *httptest.Server
	db    *DBWrapper
	ps    *providerstore.Store
	hooks engineLifecycleHooks
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
	teardownEvent := &EngineTeardownEvent{Engine: e}
	if beforeHookErr := runHooks(teardownEvent, e.hooks.beforeTeardownHooks); beforeHookErr != nil {
		return fmt.Errorf("before teardown hooks: %w", beforeHookErr)
	}

	var teardownErrs []error

	if e.ts != nil {
		e.ts.Close()
	}
	if e.db != nil && !e.db.IsZero() {
		if dbErr := e.db.Teardown(); dbErr != nil {
			teardownErrs = append(teardownErrs, dbErr)
		}
	}
	if providerErr := e.teardownProviders(); providerErr != nil {
		teardownErrs = append(teardownErrs, providerErr)
	}

	if len(teardownErrs) > 0 {
		if err := errors.Join(teardownErrs...); err != nil {
			return fmt.Errorf("failed to run engine teardown: %w", err)
		}
	}

	if afterHookErr := runHooks(teardownEvent, reverseHooks(e.hooks.afterTeardownHooks)); afterHookErr != nil {
		return fmt.Errorf("after teardown hooks: %w", afterHookErr)
	}
	return nil
}

func (e *Engine) Seed(seeds ...any) error {
	seedEvent := &EngineSeedEvent{
		Engine: e,
		Seeds:  append([]any(nil), seeds...),
	}
	if beforeHookErr := runHooks(seedEvent, e.hooks.beforeSeedHooks); beforeHookErr != nil {
		return beforeHookErr
	}

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
	runEvent := &EngineRunEvent{
		TB:     t,
		Engine: e,
		Runner: runner,
		Ctx:    ctx,
	}

	if beforeHookErr := runHooks(runEvent, e.hooks.beforeRunHooks); beforeHookErr != nil {
		return fmt.Errorf("failed to run `Engine.Execute` before hooks: %w", beforeHookErr)
	}

	if runErr := runner.Run(t, ctx); runErr != nil {
		return fmt.Errorf("failed to execute runners on Engine: %w", runErr)
	}

	if afterHookErr := runHooks(runEvent, reverseHooks(e.hooks.afterRunHooks)); afterHookErr != nil {
		return fmt.Errorf("failed to run `Engine.Execute` after hooks: %w", afterHookErr)
	}

	return nil
}
