package puppetest

import (
	"context"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

type (
	EngineExtension func(engine *Engine) error
	EngineFactory   struct {
		dbFactory  *dbastidor.ConnectionFactory
		extensions []EngineExtension
	}
	EngineFactoryOption func(*EngineFactory) error
)

func WithConnectionFactory(
	connPerformer dbastidor.ConnectionPerformer, executeDbCreateStmt bool,
) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		dbFactory, err := dbastidor.NewConnectionFactory(
			context.Background(), executeDbCreateStmt, connPerformer,
		)
		fac.dbFactory = dbFactory
		return err
	}
}

func WithExtensions(extensions ...EngineExtension) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.extensions = append(fac.extensions, extensions...)
		return nil
	}
}

func NewEngineFactory(
	options ...EngineFactoryOption,
) (*EngineFactory, error) {
	newFactory := &EngineFactory{}
	for _, opt := range options {
		if err := opt(newFactory); err != nil {
			return newFactory, err
		}
	}

	return newFactory, nil
}

func (fac EngineFactory) NewEngine(t testing.TB) *Engine {
	engine := new(Engine)
	var dbTeardown func(ctx context.Context) error
	if fac.dbFactory != nil {
		subDb, err := fac.dbFactory.NewDatabase(t.Context(), t.Name())
		if err != nil {
			t.Fatal(err)
		}
		engine.db = NewDBWrapper(subDb.Name, subDb.Connection)
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

	for _, extension := range fac.extensions {
		if err := extension(engine); err != nil {
			t.Fatal(err)
		}
	}

	return engine
}

func (fac EngineFactory) Close() error {
	if fac.dbFactory != nil {
		return fac.dbFactory.Close()
	}
	return nil
}
