package puppetest

import (
	"context"
	"errors"

	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
)

type ProviderKey = providerstore.Key

func NewProviderKey[V any](_ ...V) ProviderKey {
	return providerstore.NewKey[V]()
}

func NewTaggedProviderKey[V any](tag string, _ ...V) ProviderKey {
	return providerstore.NewTaggedKey[V](tag)
}

func (e *Engine) providerStore() *providerstore.Store {
	if e.ps == nil {
		e.ps = providerstore.New()
	}
	return e.ps
}

func SetProvider[T any](
	engine *Engine, key ProviderKey, value *T, teardown func(context.Context, *T) error,
) error {
	if engine == nil {
		return errors.New("engine is nil")
	}
	if value == nil {
		return errors.New("provider value is nil")
	}

	var internalTeardown func(context.Context) error
	if teardown != nil {
		internalTeardown = func(ctx context.Context) error {
			return teardown(ctx, value)
		}
	}

	return engine.providerStore().Save(key, value, internalTeardown)
}

func Provider[T any](engine *Engine, key ProviderKey) (*T, bool) {
	if engine == nil || engine.ps == nil {
		return nil, false
	}

	value, found := engine.ps.Load(key)
	if !found {
		return nil, false
	}

	casted, ok := value.(*T)
	if !ok || casted == nil {
		return nil, false
	}
	return casted, true
}

func (e *Engine) teardownProviders() error {
	if e == nil || e.ps == nil {
		return nil
	}
	return e.ps.Teardown(context.Background())
}
