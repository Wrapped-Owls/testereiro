package puppetest

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
)

type (
	EngineExtension func(engine *Engine) error
	EngineFactory   struct {
		dbFactory        *dbastidor.ConnectionFactory
		connPerformer    dbastidor.ConnectionPerformer
		lifecycleBuilder dbastidor.LifecycleBuilder
		placeholderStyle dbastidor.PlaceholderStyle
		ps               *providerstore.Store
		binders          map[ProviderKey]factoryProviderBinder
		extensions       []EngineExtension
		hookLifecycle    engineFactoryHookLifecycle
	}
	EngineFactoryOption func(*EngineFactory) error
)

func (fac *EngineFactory) providerStore() *providerstore.Store {
	if fac.ps == nil {
		fac.ps = providerstore.New()
	}
	return fac.ps
}

func NewEngineFactory(
	options ...EngineFactoryOption,
) (*EngineFactory, error) {
	newFactory := &EngineFactory{}
	for index, opt := range options {
		if err := opt(newFactory); err != nil {
			return nil, fmt.Errorf("apply engine factory option at index %d: %w", index, err)
		}
	}
	if err := newFactory.initConnectionFactory(context.Background()); err != nil {
		return nil, err
	}
	newFactory.hookLifecycle.bind(newFactory)

	return newFactory, nil
}

func (fac *EngineFactory) initConnectionFactory(ctx context.Context) error {
	if fac.connPerformer == nil {
		if fac.lifecycleBuilder != nil {
			return errors.New("database lifecycle set without a connection factory")
		}
		return nil
	}

	dbFactory, err := dbastidor.NewConnectionFactory(ctx, fac.connPerformer, fac.lifecycleBuilder)
	if err != nil {
		return fmt.Errorf("create connection factory: %w", err)
	}
	fac.dbFactory = dbFactory

	return nil
}

func (fac *EngineFactory) NewEngine(t testing.TB) *Engine {
	engine := new(Engine)
	if createErr := fac.hookLifecycle.handleEngineCreation(t, engine); createErr != nil {
		t.Fatal(createErr)
	}

	return engine
}

func (fac *EngineFactory) initEngine(t testing.TB, engine *Engine) error {
	var dbTeardown func(ctx context.Context) error
	engine.ctx = t.Context()
	engine.db = newDBWrapper(t.Name()+"_puppetest", nil, fac.placeholderStyle)
	if fac.dbFactory != nil {
		subDb, err := fac.dbFactory.NewDatabase(t.Context(), t.Name())
		if err != nil {
			return fmt.Errorf("create database for test %q: %w", t.Name(), err)
		}
		engine.db = newDBWrapper(subDb.Name, subDb.Connection, fac.placeholderStyle)
		dbTeardown = subDb.Teardown
	}

	t.Cleanup(
		func() {
			if dbTeardown != nil {
				t.Log("Executing database teardown")
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if err := dbTeardown(ctx); err != nil {
					t.Error(err)
				}
			}
			t.Log("Executing teardown on engine")
			shutdownErr := engine.Teardown()
			if shutdownErr != nil {
				t.Error(shutdownErr)
			}
		},
	)

	if bindErr := fac.bindFactoryProviders(t.Context(), engine); bindErr != nil {
		return fmt.Errorf("failed to bind factory providers on engine: %w", bindErr)
	}

	for _, extension := range fac.extensions {
		if err := extension(engine); err != nil {
			return fmt.Errorf("apply engine extension: %w", err)
		}
	}

	return nil
}

func (fac *EngineFactory) closeOperation() error {
	var closeErrs []error
	if fac.dbFactory != nil {
		if err := fac.dbFactory.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	if providerErr := fac.teardownProviders(context.Background()); providerErr != nil {
		closeErrs = append(closeErrs, providerErr)
	}

	return errors.Join(closeErrs...)
}

// Close is safe to call on a nil *EngineFactory.
func (fac *EngineFactory) Close() error {
	if fac == nil {
		return nil
	}

	closeErr := fac.hookLifecycle.closeFactory()
	if closeErr != nil {
		return fmt.Errorf("factory close errors: %w", closeErr)
	}
	return nil
}
