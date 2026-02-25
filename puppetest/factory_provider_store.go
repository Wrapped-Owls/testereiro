package puppetest

import (
	"context"
	"errors"
	"fmt"

	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
)

type (
	FactoryProviderBinder[T any] func(ctx context.Context, engine *Engine, value *T) error
	factoryProviderBinder        func(context.Context, *Engine) error
)

func RegisterFactoryProvider[T any](
	factory *EngineFactory,
	key ProviderKey,
	value *T,
	bind FactoryProviderBinder[T],
	teardown func(context.Context, *T) error,
) error {
	if factory == nil {
		return errors.New("engine factory is nil")
	}
	if key == nil {
		return errors.New("factory provider key is nil")
	}

	if _, exists := factory.providerStore().Load(key); exists {
		return fmt.Errorf("factory provider %s already registered", factoryProviderLabel(key))
	}

	if bind != nil {
		if factory.binders == nil {
			factory.binders = make(map[ProviderKey]factoryProviderBinder)
		}
		factory.binders[key] = func(ctx context.Context, engine *Engine) error {
			return bind(ctx, engine, value)
		}
	}

	return providerstore.SaveProvider(factory.providerStore(), key, value, teardown)
}

func FactoryProvider[T any](factory *EngineFactory, key ProviderKey) (*T, bool) {
	return ResolveProvider[T](factory, key)
}

func (fac *EngineFactory) bindFactoryProviders(ctx context.Context, engine *Engine) error {
	if fac == nil || len(fac.binders) == 0 {
		return nil
	}
	if engine == nil {
		return errors.New("engine is nil")
	}

	var bindErrs []error
	for _, key := range fac.ps.Keys() {
		bindFn, exists := fac.binders[key]
		if !exists || bindFn == nil {
			continue
		}
		if err := bindFn(ctx, engine); err != nil {
			bindErrs = append(
				bindErrs,
				fmt.Errorf("factory provider %s bind failed: %w", factoryProviderLabel(key), err),
			)
		}
	}
	return errors.Join(bindErrs...)
}

func (fac *EngineFactory) teardownProviders(ctx context.Context) error {
	if fac == nil {
		return nil
	}

	fac.binders = nil
	if fac.ps == nil {
		return nil
	}

	return fac.ps.Teardown(ctx)
}

func factoryProviderLabel(key ProviderKey) string {
	if key == nil {
		return "<nil>"
	}
	if key.Tag() != "" {
		return fmt.Sprintf("%s(%s)", key.Type(), key.Tag())
	}
	return key.Type().String()
}
