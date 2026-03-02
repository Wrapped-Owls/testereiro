package puppetest

import (
	"context"
	"errors"

	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
)

func (e *Engine) providerStore() *providerstore.Store {
	if e.ps == nil {
		e.ps = providerstore.New()
	}
	return e.ps
}

// SetProvider stores a provider value on an engine with an optional teardown callback.
func SetProvider[T any](
	engine *Engine, key ProviderKey, value *T, teardown func(context.Context, *T) error,
) error {
	if engine == nil {
		return errors.New("engine is nil")
	}
	return providerstore.SaveProvider(engine.providerStore(), key, value, teardown)
}

// Provider loads a provider value from an engine by key.
func Provider[T any](engine *Engine, key ProviderKey) (*T, bool) {
	return ResolveProvider[T](engine, key)
}

func (e *Engine) teardownProviders() error {
	ps := e.providerStore()
	return ps.Teardown(context.Background())
}
