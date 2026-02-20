package puppetest

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

type factoryProviderEntry struct {
	value    any
	bind     factoryProviderBinder
	teardown func(context.Context) error
}

type (
	FactoryProviderBinder[T any] func(
		ctx context.Context,
		factory *EngineFactory,
		engine *Engine,
		value *T,
	) error
	factoryProviderBinder func(context.Context, *EngineFactory, *Engine) error
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
	if value == nil {
		return errors.New("factory provider value is nil")
	}

	if factory.providers == nil {
		factory.providers = make(map[ProviderKey]factoryProviderEntry)
	}
	if _, exists := factory.providers[key]; exists {
		return fmt.Errorf("factory provider %s already registered", factoryProviderLabel(key))
	}

	var internalBind factoryProviderBinder
	if bind != nil {
		internalBind = func(ctx context.Context, fac *EngineFactory, engine *Engine) error {
			return bind(ctx, fac, engine, value)
		}
	}

	var internalTeardown func(context.Context) error
	if teardown != nil {
		internalTeardown = func(ctx context.Context) error {
			return teardown(ctx, value)
		}
	}

	factory.providers[key] = factoryProviderEntry{
		value:    value,
		bind:     internalBind,
		teardown: internalTeardown,
	}
	factory.providerBinderOrder = append(factory.providerBinderOrder, key)
	return nil
}

func FactoryProvider[T any](factory *EngineFactory, key ProviderKey) (*T, bool) {
	if factory == nil || key == nil || factory.providers == nil {
		return nil, false
	}

	entry, found := factory.providers[key]
	if !found || entry.value == nil {
		return nil, false
	}

	casted, ok := entry.value.(*T)
	if !ok || casted == nil {
		return nil, false
	}
	return casted, true
}

func (fac *EngineFactory) bindFactoryProviders(ctx context.Context, engine *Engine) error {
	if fac == nil || len(fac.providerBinderOrder) == 0 {
		return nil
	}
	if engine == nil {
		return errors.New("engine is nil")
	}

	var bindErrs []error
	for _, key := range fac.providerBinderOrder {
		entry, exists := fac.providers[key]
		if !exists || entry.bind == nil {
			continue
		}
		if err := entry.bind(ctx, fac, engine); err != nil {
			bindErrs = append(
				bindErrs,
				fmt.Errorf("factory provider %s bind failed: %w", factoryProviderLabel(key), err),
			)
		}
	}
	return errors.Join(bindErrs...)
}

func (fac *EngineFactory) teardownProviders(ctx context.Context) error {
	if fac == nil || len(fac.providerBinderOrder) == 0 {
		return nil
	}

	var teardownErrs []error
	for _, key := range slices.Backward(fac.providerBinderOrder) {
		entry, exists := fac.providers[key]
		if !exists || entry.teardown == nil {
			continue
		}
		if err := entry.teardown(ctx); err != nil {
			teardownErrs = append(
				teardownErrs, fmt.Errorf(
					"factory provider %s teardown failed: %w",
					factoryProviderLabel(key), err,
				),
			)
		}
	}

	fac.providers = nil
	fac.providerBinderOrder = nil

	return errors.Join(teardownErrs...)
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
