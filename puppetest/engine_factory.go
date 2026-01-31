package puppetest

import (
	"context"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

type (
	EngineExtension func(engine *Engine) error
	EngineFactory   struct {
		dbFactory  dbastidor.ConnectionFactory
		extensions []EngineExtension
	}
)

func NewEngineFactory(
	connPerformer dbastidor.ConnectionPerformer, extensions ...EngineExtension,
) (*EngineFactory, error) {
	dbFactory, err := dbastidor.NewConnectionFactory(context.Background(), false, connPerformer)
	newFactory := &EngineFactory{
		dbFactory:  dbFactory,
		extensions: extensions,
	}

	return newFactory, err
}

func (fac EngineFactory) NewEngine(t testing.TB) *Engine {
	subDb, err := fac.dbFactory.NewDatabase(t.Context(), t.Name())
	if err != nil {
		t.Fatal(err)
	}
	engine := &Engine{
		db: NewRootDBWrapper(subDb.Name, subDb.Connection),
	}

	t.Cleanup(
		func() {
			t.Log("Executing teardown on engine")
			shutdownErr := engine.Teardown()
			if shutdownErr != nil {
				t.Error(shutdownErr)
			}
		},
	)

	for _, extension := range fac.extensions {
		if err = extension(engine); err != nil {
			t.Fatal(err)
		}
	}

	return engine
}

func (fac EngineFactory) Close() error {
	return fac.dbFactory.Close()
}
