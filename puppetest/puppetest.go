package puppetest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores"
)

// Context is the internal context object used on the test engine to take some objects from a given state
type (
	Context = stgctx.RunnerContext
	// SeedProvider defines a provider-backed seed operation.
	SeedProvider interface {
		// ExecuteSeed applies data setup against the given engine.
		ExecuteSeed(engine *Engine) error
	}
)

// SaveOnCtx stores a value in the runner context using its type as key.
func SaveOnCtx[V any](ctx Context, val V) {
	stgctx.SaveOnCtx(ctx, val)
}

// LoadFromCtx retrieves a value from the runner context by type.
func LoadFromCtx[V any](ctx Context) (V, bool) {
	return stgctx.LoadFromCtx[V](ctx)
}

// Engine is the runtime test container with optional HTTP server, database, and providers.
type Engine struct {
	ctx   context.Context
	ts    *httptest.Server
	db    *DBWrapper
	ps    *providerstore.Store
	hooks engineLifecycleHooks
}

// BaseURL returns the test server URL, or an empty string when no server is configured.
func (e *Engine) BaseURL() string {
	if e.ts != nil {
		return e.ts.URL
	}
	return "" // TODO: Check a way to have this URL linked on the engine directly
}

// DB returns the engine database connection, or nil when none is configured.
func (e *Engine) DB() *sql.DB {
	if e.db == nil {
		return nil
	}
	return e.db.Connection()
}

// DBName returns the normalized engine database name, or an empty string when unavailable.
func (e *Engine) DBName() string {
	if e.db == nil {
		return ""
	}

	return e.db.name
}

// Context returns the engine context or context.Background when unset.
func (e *Engine) Context() context.Context {
	if e.ctx == nil {
		return context.Background()
	}

	return e.ctx
}

// Teardown closes engine resources and runs teardown hooks.
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

// Seed executes SQL-based seed structs against the engine database.
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

// SeedWithProvider executes provider-backed seeding strategies.
func (e *Engine) SeedWithProvider(providers ...SeedProvider) error {
	seedEvent := &EngineSeedEvent{
		Engine:        e,
		ProviderSeeds: slices.Clone(providers),
	}
	if beforeHookErr := runHooks(seedEvent, e.hooks.beforeSeedHooks); beforeHookErr != nil {
		return beforeHookErr
	}

	var seedErrs []error
	for index, provider := range providers {
		if provider == nil {
			seedErrs = append(seedErrs, fmt.Errorf("seed provider at index %d is nil", index))
			continue
		}
		if err := provider.ExecuteSeed(e); err != nil {
			seedErrs = append(
				seedErrs,
				fmt.Errorf("seed provider at index %d failed: %w", index, err),
			)
		}
	}
	return errors.Join(seedErrs...)
}

// Execute runs a runner with hook lifecycle and shared runner context.
func (e *Engine) Execute(t testing.TB, runner atores.Runner) error {
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
